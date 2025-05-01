package table

import (
	"github.com/xhd2015/ormx/field"
)

// Table represents a database table
type Table struct {
	name string
}

// New creates a new Table
func New(name string) Table {
	return Table{name: name}
}

// Name returns the table name
func (t Table) Name() string {
	return t.name
}

// Int64 creates a new Int64Field for this table
func (t Table) Int64(name string) field.Int64Field {
	return field.Int64Field{
		FieldName: name,
		TableName: t.name,
	}
}

// String creates a new StringField for this table
func (t Table) String(name string) field.StringField {
	return field.StringField{
		FieldName: name,
		TableName: t.name,
	}
}

// Time creates a new TimeField for this table
func (t Table) Time(name string) field.TimeField {
	return field.TimeField{
		FieldName: name,
		TableName: t.name,
	}
}
