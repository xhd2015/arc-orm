package field

import "time"

// TimeField represents a time.Time database field
type TimeField struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f TimeField) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f TimeField) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f TimeField) ToSQL() string {
	return "`" + f.TableName + "`.`" + f.FieldName + "`"
}

// Eq creates an equality condition (field = value)
func (f TimeField) Eq(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f TimeField) EqField(other Field) Condition {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f TimeField) Neq(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

// Gt creates a greater than condition (field > value)
func (f TimeField) Gt(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

// Gte creates a greater than or equal condition (field >= value)
func (f TimeField) Gte(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

// Lt creates a less than condition (field < value)
func (f TimeField) Lt(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

// Lte creates a less than or equal condition (field <= value)
func (f TimeField) Lte(value time.Time) Condition {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

// Between creates a BETWEEN condition
func (f TimeField) Between(start, end time.Time) Condition {
	return &between{
		field: f,
		start: start,
		end:   end,
	}
}

// Asc returns an ascending order specification for this field
func (f TimeField) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f TimeField) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f TimeField) As(alias string) Field {
	return As(f, alias)
}
