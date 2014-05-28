package gorethink

import (
	"sync"
	"sync/atomic"
	"time"

	"code.google.com/p/goprotobuf/proto"
	p "github.com/dancannon/gorethink/ql2"
)

type Session struct {
	token      int64
	address    string
	database   string
	timeout    time.Duration
	authkey    string
	timeFormat string

	// Pool configuration options
	maxIdle     int
	maxActive   int
	idleTimeout time.Duration

	// Response cache, used for batched responses
	sync.Mutex
	cache map[int64]*ResultRows

	closed bool

	pool *Pool
}

func newSession(args map[string]interface{}) *Session {
	s := &Session{
		cache: map[int64]*ResultRows{},
	}

	if token, ok := args["token"]; ok {
		s.token = token.(int64)
	}
	if address, ok := args["address"]; ok {
		s.address = address.(string)
	}
	if database, ok := args["database"]; ok {
		s.database = database.(string)
	}
	if timeout, ok := args["timeout"]; ok {
		s.timeout = timeout.(time.Duration)
	}
	if authkey, ok := args["authkey"]; ok {
		s.authkey = authkey.(string)
	}

	// Pool configuration options
	if maxIdle, ok := args["maxIdle"]; ok {
		s.maxIdle = maxIdle.(int)
	} else {
		s.maxIdle = 1
	}
	if maxActive, ok := args["maxActive"]; ok {
		s.maxActive = maxActive.(int)
	} else {
		s.maxActive = 0
	}
	if idleTimeout, ok := args["idleTimeout"]; ok {
		s.idleTimeout = idleTimeout.(time.Duration)
	} else {
		s.idleTimeout = 10 * time.Second
	}

	return s
}

type ConnectOpts struct {
	Token       int64         `gorethink:"token,omitempty"`
	Address     string        `gorethink:"address,omitempty"`
	Database    string        `gorethink:"database,omitempty"`
	Timeout     time.Duration `gorethink:"timeout,omitempty"`
	AuthKey     string        `gorethink:"authkey,omitempty"`
	MaxIdle     int           `gorethink:"max_idle,omitempty"`
	MaxActive   int           `gorethink:"max_active,omitempty"`
	IdleTimeout time.Duration `gorethink:"idle_timeout,omitempty"`
}

func (o *ConnectOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Connect creates a new database session.
//
// Supported arguments include token, address, database, timeout, authkey,
// and timeFormat. Pool options include maxIdle, maxActive and idleTimeout.
//
// By default maxIdle and maxActive are set to 1: passing values greater
// than the default (e.g. maxIdle: "10", maxActive: "20") will provide a
// pool of re-usable connections.
//
// Basic connection example:
//
//	var session *r.Session
// 	session, err := r.Connect(r.ConnectOpts{
// 		Address:  "localhost:28015",
// 		Database: "test",
// 		AuthKey:  "14daak1cad13dj",
// 	})
func Connect(args ConnectOpts) (*Session, error) {
	s := newSession(args.toMap())
	err := s.Reconnect()

	return s, err
}

type CloseOpts struct {
	NoReplyWait bool `gorethink:"noreplyWait,omitempty"`
}

func (o *CloseOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Reconnect closes and re-opens a session.
func (s *Session) Reconnect(optArgs ...CloseOpts) error {
	if err := s.Close(optArgs...); err != nil {
		return err
	}

	s.closed = false
	if s.pool == nil {
		s.pool = &Pool{
			Session:     s,
			MaxIdle:     s.maxIdle,
			MaxActive:   s.maxActive,
			IdleTimeout: s.idleTimeout,
		}
	}

	// Check the connection
	conn, err := s.pool.get()
	if err == nil {
		conn.Close()
	}

	return err
}

// Close closes the session
func (s *Session) Close(optArgs ...CloseOpts) error {
	if s.closed {
		return nil
	}

	if len(optArgs) >= 1 {
		if optArgs[0].NoReplyWait {
			s.NoReplyWait()
		}
	}

	var err error
	if s.pool != nil {
		err = s.pool.Close()
	}
	s.closed = true

	return err
}

// noreplyWait ensures that previous queries with the noreply flag have been
// processed by the server. Note that this guarantee only applies to queries
// run on the given connection
func (s *Session) NoReplyWait() {
	s.noreplyWaitQuery()
}

// Use changes the default database used
func (s *Session) Use(database string) {
	s.database = database
}

// SetTimeout causes any future queries that are run on this session to timeout
// after the given duration, returning a timeout error.  Set to zero to disable.
func (s *Session) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns
// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit
//
// If n <= 0, no idle connections are retained.
func (s *Session) SetMaxIdleConns(n int) {
	s.pool.MaxIdle = n
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
//
// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
// MaxIdleConns, then MaxIdleConns will be reduced to match the new
// MaxOpenConns limit
//
// If n <= 0, then there is no limit on the number of open connections.
// The default is 0 (unlimited).
func (s *Session) SetMaxOpenConns(n int) {
	s.pool.MaxActive = n
}

// getToken generates the next query token, used to number requests and match
// responses with requests.
func (s *Session) nextToken() int64 {
	return atomic.AddInt64(&s.token, 1)
}

// startQuery creates a query from the term given and sends it to the server.
// The result from the server is returned as ResultRows
func (s *Session) startQuery(t RqlTerm, opts map[string]interface{}) (*ResultRows, error) {
	token := s.nextToken()

	// Build query tree
	pt := t.build()

	// Build global options
	globalOpts := []*p.Query_AssocPair{}
	for k, v := range opts {
		globalOpts = append(globalOpts, &p.Query_AssocPair{
			Key: proto.String(k),
			Val: Expr(v).build(),
		})
	}

	// If no DB option was set default to the value set in the connection
	if _, ok := opts["db"]; !ok {
		globalOpts = append(globalOpts, &p.Query_AssocPair{
			Key: proto.String("db"),
			Val: Db(s.database).build(),
		})
	}

	// Construct query
	q := &p.Query{
		AcceptsRJson:  proto.Bool(true),
		Type:          p.Query_START.Enum(),
		Token:         proto.Int64(token),
		Query:         pt,
		GlobalOptargs: globalOpts,
	}

	conn := s.pool.Get()
	defer conn.Close()

	return conn.SendQuery(s, q, t, opts, false)
}

func (s *Session) handleBatchResponse(response *p.Response) {
	s.Lock()
	result := s.cache[response.GetToken()]
	s.Unlock()

	result.extend(response)
	result.outstandingRequests--

	if response.GetType() != p.Response_SUCCESS_PARTIAL && result.outstandingRequests == 0 {
		s.Lock()
		delete(s.cache, response.GetToken())
		s.Unlock()
	}
}

// continueQuery continues a previously run query.
// This is needed if a response is batched.
func (s *Session) continueQuery(result *ResultRows) error {
	s.Lock()
	s.cache[result.query.GetToken()].outstandingRequests++
	s.Unlock()

	conn := s.pool.Get()
	defer conn.Close()

	q := &p.Query{
		Type:  p.Query_CONTINUE.Enum(),
		Token: result.query.Token,
	}

	_, err := conn.SendQuery(s, q, result.term, result.opts, true)
	if err != nil {
		return err
	}

	response, err := conn.ReadResponse(s, result.query.GetToken())
	if err != nil {
		return err
	}

	s.handleBatchResponse(response)

	return nil
}

// asyncContinueQuery asynchronously continues a previously run query.
// This is needed if a response is batched.
func (s *Session) asyncContinueQuery(result *ResultRows) error {
	s.Lock()
	s.cache[result.query.GetToken()].outstandingRequests++
	s.Unlock()

	conn := s.pool.Get()
	defer conn.Close()

	q := &p.Query{
		Type:  p.Query_CONTINUE.Enum(),
		Token: result.query.Token,
	}

	_, err := conn.SendQuery(s, q, result.term, result.opts, true)
	if err != nil {
		return err
	}

	return nil
}

// stopQuery sends closes a query by sending Query_STOP to the server.
func (s *Session) stopQuery(result *ResultRows) error {
	q := &p.Query{
		Type:  p.Query_STOP.Enum(),
		Token: result.query.Token,
	}

	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.SendQuery(s, q, result.term, result.opts, false)
	if err != nil {
		return err
	}

	response, err := conn.ReadResponse(s, result.query.GetToken())
	if err != nil {
		return err
	}

	s.handleBatchResponse(response)

	return nil
}

// noreplyWaitQuery sends the NOREPLY_WAIT query to the server.
func (s *Session) noreplyWaitQuery() error {
	q := &p.Query{
		Type:  p.Query_NOREPLY_WAIT.Enum(),
		Token: proto.Int64(s.nextToken()),
	}

	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.SendQuery(s, q, RqlTerm{}, map[string]interface{}{}, false)

	return err
}
