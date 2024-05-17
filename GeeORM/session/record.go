package session

import (
	"geeorm/clause"
	"reflect"
)

func (s *Session) Where(desc string, v ...interface{}) *Session {
	//
	return nil
}
func (s *Session) Orderby(desc string) *Session {
	return nil
}
func (s *Session) Find(value interface{}) error {
	dstSlice := reflect.Indirect(reflect.ValueOf(value))
	dstType := dstSlice.Type().Elem()
	table := s.Model(reflect.New(dstType).Elem().Interface()).refTable
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	for rows.Next() {
		dest := reflect.New(dstType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		dstSlice.Set(reflect.Append(dstSlice, dest))
	}
	return rows.Close()
}
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.Model(value).refTable
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Update(input ...interface{}) (int64, error) {
	return 0, nil
}
func (s *Session) Delete() (int64, error) {
	return 0, nil
}
