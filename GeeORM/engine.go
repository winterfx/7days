package geeORM

import (
	"database/sql"
	"geeorm/session"
)

type engine interface {
	NewSession() *session.Session
	Close() error
}

type ormEngine struct {
	db *sql.DB
}

func (e *ormEngine) NewSession() *session.Session {
	return session.NewSession(e.db)
}
func (e *ormEngine) Close() error {
	return e.db.Close()
}
func NewEngine(driver, source string) engine {
	db, err := sql.Open(driver, source)
	if err != nil {
		panic(err.Error())
	}
	if err = db.Ping(); err != nil {
		panic(err.Error())
	}
	return &ormEngine{db: db}
}
