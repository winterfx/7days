// Package dialect define sql interface to shield databases data type differences(sqlite/mysql)
package dialect

import "reflect"

// 为什么不需要定一个接口里面是数据库的crud,因为跟数据库驱动交互的最终形态就是sql语句，只需要生成不同的sql语句即可，不同数据库也是数据类型的不同
type Dialect interface {
	//cover data object to data
	DataTypeOf(typ reflect.Value) string
	//TableExistSQL(tableName string)(string,)
}
