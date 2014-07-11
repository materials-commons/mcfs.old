package model

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
)

// ModelQueries holds the different types of sql queries to perform
// tasks such as insert on a model
type ModelQueries struct {
	Insert string
}

// Model describes the data model for a table/object.
type Model struct {
	schema  interface{}
	table   string
	typeOf  reflect.Type
	ptrOf   reflect.Type
	queries ModelQueries
}

// Query holds the model and the connection to the database for
// performing queries against that model.
type Query struct {
	*Model
	*sqlx.DB
}

// New creates a new Model for a specific table, along with the CRUD queries
// for interacting with that model.
func New(schema interface{}, table string, mq ModelQueries) *Model {
	typeOf := reflect.TypeOf(schema)
	return &Model{
		schema:  schema,
		table:   table,
		typeOf:  typeOf,
		ptrOf:   reflect.PtrTo(typeOf),
		queries: mq,
	}
}

// Table returns the name of the table for a model.
func (m *Model) Table() string {
	return m.table
}

// Q takes a database connection to use for a query.
func (m *Model) Q(db *sqlx.DB) *Query {
	return &Query{
		Model: m,
		DB:    db,
	}
}

// ByID queries a model by its primary (integer) key.
func (q *Query) ByID(id int) (interface{}, error) {
	result := reflect.New(reflect.TypeOf(q.schema))
	query := fmt.Sprintf("select * from %s where id = ?", q.table)
	err := q.Get(result.Interface(), query, id)
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

// T takes a query with a %s token in it and adds the table. This
// method allows us to divorce a query from the models table.
func (m *Model) T(query string) string {
	return fmt.Sprint(query, m.table)
}

// Insert performs a database insert.
func (q *Query) Insert(item interface{}) error {
	t := reflect.TypeOf(item)
	switch {
	case t == q.typeOf:
	case t == q.ptrOf:
	default:
		return fmt.Errorf("wrong type for model")
	}

	_, err := q.NamedExec(q.queries.Insert, item)
	return err
}
