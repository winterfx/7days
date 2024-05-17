package session

import (
	"database/sql"
	"geeorm/clause"
	"geeorm/dialect"
	"geeorm/schema"
	"strings"
)

type commonDB interface {
	Exec() (sql.Result, error)
	QueryRow() *sql.Row
	QueryRows() (*sql.Rows, error)
}

type Session struct {
	db       *sql.DB
	sql      strings.Builder
	sqlVar   []interface{}
	refTable *schema.Schema
	dialect  dialect.Dialect
	clause   *clause.Clause
	commonDB
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVar = nil
}
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVar = append(s.sqlVar, values...)
	return s
}
func (s *Session) Exec() (sql.Result, error) {
	defer s.Clear()
	return s.db.Exec(s.sql.String(), s.sqlVar...)
}
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	return s.db.QueryRow(s.sql.String(), s.sqlVar...)
}
func (s *Session) QueryRows() (*sql.Rows, error) {
	defer s.Clear()
	return s.db.Query(s.sql.String(), s.sqlVar...)
}
func NewSession(db *sql.DB) *Session {
	return &Session{
		db: db,
	}
}
