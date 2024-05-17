package clause

import "strings"

type Type int

const (
	INSERT Type = iota + 1
	VALUES
	SELECT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
	LIMIT
)

type Clause struct {
	sql    map[Type]string
	sqlVar map[Type][]interface{}
}

// first set and build
func (c *Clause) Set(typ Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVar = make(map[Type][]interface{})
	}
	sql, sqlVar := generators[typ](vars...)
	c.sql[typ] = sql
	c.sqlVar[typ] = sqlVar
}
func (c *Clause) Build(order ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, t := range order {
		if sql, ok := c.sql[t]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVar[t]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
