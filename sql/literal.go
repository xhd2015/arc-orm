package sql

import "time"

// String is a string literal expression for use in SQL statements
type String string

// ToExpressionSQL returns the SQL for a string literal
func (s String) ToExpressionSQL() (string, interface{}) {
	return "", string(s)
}

// Int64 is an int64 literal expression for use in SQL statements
type Int64 int64

// ToExpressionSQL returns the SQL for an int64 literal
func (i Int64) ToExpressionSQL() (string, interface{}) {
	return "", int64(i)
}

type Int32 int32

// ToExpressionSQL returns the SQL for an int32 literal
func (i Int32) ToExpressionSQL() (string, interface{}) {
	return "", int32(i)
}

// Float64 is a float64 literal expression for use in SQL statements
type Float64 float64

// ToExpressionSQL returns the SQL for a float64 literal
func (f Float64) ToExpressionSQL() (string, interface{}) {
	return "", float64(f)
}

// Bool is a boolean literal expression for use in SQL statements
type Bool bool

// ToExpressionSQL returns the SQL for a boolean literal
func (b Bool) ToExpressionSQL() (string, interface{}) {
	return "", bool(b)
}

// Time is a time.Time literal expression for use in SQL statements
type Time time.Time

// ToExpressionSQL returns the SQL for a time literal
func (t Time) ToExpressionSQL() (string, interface{}) {
	return "", time.Time(t)
}
