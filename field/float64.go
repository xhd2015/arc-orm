package field

// Float64Field represents a float64 database field
type Float64Field struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f Float64Field) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f Float64Field) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f Float64Field) ToSQL() string {
	// If the field has no table, or the table name is empty, just use the field name
	if f.TableName == "" {
		return "`" + f.FieldName + "`"
	}
	return "`" + f.TableName + "`.`" + f.FieldName + "`"
}

// Eq creates an equality condition (field = value)
func (f Float64Field) Eq(value float64) Condition {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f Float64Field) EqField(other Field) Condition {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f Float64Field) Neq(value float64) Condition {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

// Gt creates a greater than condition (field > value)
func (f Float64Field) Gt(value float64) Condition {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

// Gte creates a greater than or equal condition (field >= value)
func (f Float64Field) Gte(value float64) Condition {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

// Lt creates a less than condition (field < value)
func (f Float64Field) Lt(value float64) Condition {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

// Lte creates a less than or equal condition (field <= value)
func (f Float64Field) Lte(value float64) Condition {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

// In creates an IN condition (field IN (values))
func (f Float64Field) In(values ...float64) Condition {
	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}
	return &inCondition{
		field:  f,
		values: interfaceValues,
	}
}

// Asc returns an ascending order specification for this field
func (f Float64Field) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f Float64Field) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f Float64Field) As(alias string) Field {
	return As(f, alias)
}

// Increment returns an expression to increment this field by a value
func (f Float64Field) Increment(value float64) Expression {
	return &fieldOperation{
		field:    f,
		operator: "+",
		value:    value,
	}
}

// Decrement returns an expression to decrement this field by a value
func (f Float64Field) Decrement(value float64) Expression {
	return &fieldOperation{
		field:    f,
		operator: "-",
		value:    value,
	}
}
