package gorethink

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"code.google.com/p/goprotobuf/proto"
	p "github.com/dancannon/gorethink/ql2"
)

type Conn interface {
	SendQuery(s *Session, q *p.Query, t RqlTerm, opts map[string]interface{}, async bool) (*ResultRows, error)
	ReadResponse(s *Session, token int64) (*p.Response, error)
	Close() error
}

// connection is a connection to a rethinkdb database
type Connection struct {
	// embed the net.Conn type, so that we can effectively define new methods on
	// it (interfaces do not allow that)
	net.Conn

	closed bool
}

// Reconnect closes the previous connection and attempts to connect again.
func Dial(s *Session) (*Connection, error) {
	conn, err := net.Dial("tcp", s.address)
	if err != nil {
		return nil, RqlConnectionError{err.Error()}
	}

	if err := binary.Write(conn, binary.LittleEndian, p.VersionDummy_V0_2); err != nil {
		return nil, RqlConnectionError{err.Error()}
	}

	// authorization key
	if err := binary.Write(conn, binary.LittleEndian, uint32(len(s.authkey))); err != nil {
		return nil, RqlConnectionError{err.Error()}
	}

	if err := binary.Write(conn, binary.BigEndian, []byte(s.authkey)); err != nil {
		return nil, RqlConnectionError{err.Error()}
	}

	// read server response to authorization key (terminated by NUL)
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\x00')
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("Unexpected EOF: %s", string(line))
		}
		return nil, RqlDriverError{err.Error()}
	}
	// convert to string and remove trailing NUL byte
	response := string(line[:len(line)-1])
	if response != "SUCCESS" {
		// we failed authorization or something else terrible happened
		return nil, RqlDriverError{fmt.Sprintf("Server dropped connection with message: \"%s\"", response)}
	}

	return &Connection{conn, false}, nil
}

func TestOnBorrow(c *Connection, t time.Time) error {
	c.SetReadDeadline(t)

	data := make([]byte, 1)
	if _, err := c.Read(data); err != nil {
		e, ok := err.(net.Error)
		if err != nil && !(ok && e.Timeout()) {
			return err
		}
	}

	c.SetReadDeadline(time.Time{})
	return nil
}

func (c *Connection) ReadResponse(s *Session, token int64) (*p.Response, error) {
	for {
		var messageLength uint32
		if err := binary.Read(c, binary.LittleEndian, &messageLength); err != nil {
			c.Close()
			return nil, RqlConnectionError{err.Error()}
		}

		buffer := make([]byte, messageLength)
		if _, err := io.ReadFull(c, buffer); err != nil {
			c.Close()
			return nil, RqlDriverError{err.Error()}
		}

		response := &p.Response{}
		if err := proto.Unmarshal(buffer, response); err != nil {
			return nil, RqlDriverError{err.Error()}
		}

		if response.GetToken() == token {
			return response, nil
		} else if _, ok := s.cache[response.GetToken()]; ok {
			// Handle batch response
			s.handleBatchResponse(response)
		} else {
			return nil, RqlDriverError{"Unexpected response received"}
		}
	}
}

func (c *Connection) SendQuery(s *Session, q *p.Query, t RqlTerm, opts map[string]interface{}, async bool) (*ResultRows, error) {
	var data []byte
	var err error

	// Ensure that the connection is not closed
	if s.closed {
		return nil, RqlDriverError{"Connection is closed"}
	}

	// Set timeout
	if s.timeout == 0 {
		c.SetDeadline(time.Time{})
	} else {
		c.SetDeadline(time.Now().Add(s.timeout))
	}

	// Send query
	if data, err = proto.Marshal(q); err != nil {
		return nil, RqlDriverError{err.Error()}
	}
	if err = binary.Write(c, binary.LittleEndian, uint32(len(data))); err != nil {
		c.Close()
		return nil, RqlConnectionError{err.Error()}
	}

	if err = binary.Write(c, binary.BigEndian, data); err != nil {
		c.Close()
		return nil, RqlConnectionError{err.Error()}
	}

	// Return immediately if the noreply option was set
	if noreply, ok := opts["noreply"]; ok && noreply.(bool) {
		return nil, nil
	} else if async {
		return nil, nil
	}

	// Get response
	response, err := c.ReadResponse(s, *q.Token)
	if err != nil {
		return nil, err
	}

	err = checkErrorResponse(response, t)
	if err != nil {
		return nil, err
	}

	// De-construct the profile datum if it exists
	var profile interface{}
	if response.GetProfile() != nil {
		var err error

		profile, err = deconstructDatum(response.GetProfile(), opts)
		if err != nil {
			return nil, RqlDriverError{err.Error()}
		}
	}

	// De-construct datum and return the result
	switch response.GetType() {
	case p.Response_SUCCESS_PARTIAL, p.Response_SUCCESS_SEQUENCE:
		result := &ResultRows{
			session: s,
			query:   q,
			term:    t,
			opts:    opts,
			profile: profile,
		}

		s.Lock()
		s.cache[*q.Token] = result
		s.Unlock()

		result.extend(response)

		return result, nil
	case p.Response_SUCCESS_ATOM:
		var value []interface{}
		var err error

		if len(response.GetResponse()) < 1 {
			value = []interface{}{}
		} else if response.GetResponse()[0].GetType() == p.Datum_R_ARRAY {
			value, err = deconstructDatums(response.GetResponse()[0].GetRArray(), opts)
			if err != nil {
				return nil, err
			}
		} else {
			var v interface{}

			v, err = deconstructDatum(response.GetResponse()[0], opts)
			if err != nil {
				return nil, RqlDriverError{err.Error()}
			}

			if sv, ok := v.([]interface{}); ok {
				value = sv
			} else if v == nil {
				value = []interface{}{nil}
			} else {
				value = []interface{}{v}
			}
		}

		return &ResultRows{
			session:  s,
			query:    q,
			term:     t,
			opts:     opts,
			profile:  profile,
			buffer:   value,
			finished: true,
		}, nil
	case p.Response_WAIT_COMPLETE:
		return nil, nil
	default:
		return nil, RqlDriverError{fmt.Sprintf("Unexpected response type received: %s", response.GetType())}
	}
}

func (c *Connection) Close() error {
	c.closed = true
	return c.Conn.Close()
}

func checkErrorResponse(response *p.Response, t RqlTerm) error {
	switch response.GetType() {
	case p.Response_CLIENT_ERROR:
		return RqlClientError{rqlResponseError{response, t}}
	case p.Response_COMPILE_ERROR:
		return RqlCompileError{rqlResponseError{response, t}}
	case p.Response_RUNTIME_ERROR:
		return RqlRuntimeError{rqlResponseError{response, t}}
	}

	return nil
}
