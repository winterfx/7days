package session

import (
	"fmt"
	"geeorm/schema"
	"reflect"
	"strings"
)

func (s *Session) Model(input interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(input) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Convert(input, s.dialect)
	}
	return s
}
func (s *Session) CreateTable() error {
	var columns []string
	for _, field := range s.refTable.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABEL %s (%s);", s.refTable.Name, desc)).Exec()
	return err
}
func (s *Session) DropTable() error {
	return nil
}
func (s *Session) HasTable() bool {
	return false
}
