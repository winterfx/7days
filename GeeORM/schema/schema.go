package schema

import (
	"geeorm/dialect"
	"reflect"
)

type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fileMap    map[string]*Field
}

// Values return the values of dest's member variables
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
func Convert(input interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(input)).Type()
	schema := &Schema{
		Model:      input,
		Name:       modelType.Name(),
		Fields:     make([]*Field, 0),
		fileMap:    make(map[string]*Field),
		FieldNames: make([]string, 0),
	}
	for i := 0; i < modelType.NumField(); i++ {
		fs := modelType.Field(i)
		f := &Field{
			Name: fs.Name,
			Type: d.DataTypeOf(reflect.Indirect(reflect.New(fs.Type))),
		}
		if v, ok := fs.Tag.Lookup("geeorm"); ok {
			f.Tag = v
		}
		schema.Fields = append(schema.Fields, f)
		schema.FieldNames = append(schema.FieldNames, f.Name)
		schema.fileMap[f.Name] = f
	}
	return schema
}
