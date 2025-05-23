package table

import (
	"github.com/xhd2015/arc-orm/field"
)

// Table represents a database table
type Table struct {
	name   string
	fields []field.Field
}

// New creates a new Table
func New(name string) Table {
	return Table{
		name:   name,
		fields: make([]field.Field, 0),
	}
}

// Name returns the table name
func (t Table) Name() string {
	return t.name
}

// Fields returns all fields associated with this table
func (t Table) Fields() []field.Field {
	return t.fields
}

// Int64 creates a new Int64Field for this table
func (t *Table) Int64(name string) field.Int64Field {
	f := field.Int64Field{
		FieldName: name,
		TableName: t.name,
	}
	t.fields = append(t.fields, f)
	return f
}

// Int32 creates a new Int32Field for this table
func (t *Table) Int32(name string) field.Int32Field {
	f := field.Int32Field{
		FieldName: name,
		TableName: t.name,
	}
	t.fields = append(t.fields, f)
	return f
}

// Float64 creates a new Float64Field for this table
func (t *Table) Float64(name string) field.Float64Field {
	f := field.Float64Field{
		FieldName: name,
		TableName: t.name,
	}
	t.fields = append(t.fields, f)
	return f
}

// String creates a new StringField for this table
func (t *Table) String(name string) field.StringField {
	f := field.StringField{
		FieldName: name,
		TableName: t.name,
	}
	t.fields = append(t.fields, f)
	return f
}

// Time creates a new TimeField for this table
func (t *Table) Time(name string) field.TimeField {
	f := field.TimeField{
		FieldName: name,
		TableName: t.name,
	}
	t.fields = append(t.fields, f)
	return f
}
