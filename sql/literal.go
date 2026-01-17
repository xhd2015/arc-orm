package sql

import "time"

// String is a string literal expression for use in SQL statements
type String string

// ToSQL implements field.Expr for string literals
func (s String) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{string(s)}, nil
}

// Int64 is an int64 literal expression for use in SQL statements
type Int64 int64

// ToSQL implements field.Expr for int64 literals
func (i Int64) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{int64(i)}, nil
}

type Int32 int32

// ToSQL implements field.Expr for int32 literals
func (i Int32) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{int32(i)}, nil
}

// Float64 is a float64 literal expression for use in SQL statements
type Float64 float64

// ToSQL implements field.Expr for float64 literals
func (f Float64) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{float64(f)}, nil
}

// Bool is a boolean literal expression for use in SQL statements
type Bool bool

// ToSQL implements field.Expr for boolean literals
func (b Bool) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{bool(b)}, nil
}

// Time is a time.Time literal expression for use in SQL statements
type Time time.Time

// ToSQL implements field.Expr for time literals
func (t Time) ToSQL() (string, []interface{}, error) {
	return "?", []interface{}{time.Time(t)}, nil
}
