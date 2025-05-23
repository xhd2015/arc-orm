package field

// Int64Field represents an int64 database field
type Int64Field struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f Int64Field) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f Int64Field) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f Int64Field) ToSQL() string {
	// If the field has no table, or the table name is empty, just use the field name
	if f.TableName == "" {
		return "`" + f.FieldName + "`"
	}
	return "`" + f.TableName + "`.`" + f.FieldName + "`"
}

// Eq creates an equality condition (field = value)
func (f Int64Field) Eq(value int64) Condition {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f Int64Field) EqField(other Field) Condition {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f Int64Field) Neq(value int64) Condition {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

// Gt creates a greater than condition (field > value)
func (f Int64Field) Gt(value int64) Condition {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

// Gte creates a greater than or equal condition (field >= value)
func (f Int64Field) Gte(value int64) Condition {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

// Lt creates a less than condition (field < value)
func (f Int64Field) Lt(value int64) Condition {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

// Lte creates a less than or equal condition (field <= value)
func (f Int64Field) Lte(value int64) Condition {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

// IsNull creates an IS NULL condition (field IS NULL)
func (f Int64Field) IsNull() Condition {
	return &nullCondition{
		field:  f,
		isNull: true,
	}
}

// In creates an IN condition (field IN (values))
func (f Int64Field) In(values ...int64) Condition {
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
func (f Int64Field) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f Int64Field) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f Int64Field) As(alias string) Field {
	return As(f, alias)
}

// Increment returns an expression to increment this field by a value
func (f Int64Field) Increment(value int64) Expression {
	return &fieldOperation{
		field:    f,
		operator: "+",
		value:    value,
	}
}

// Decrement returns an expression to decrement this field by a value
func (f Int64Field) Decrement(value int64) Expression {
	return &fieldOperation{
		field:    f,
		operator: "-",
		value:    value,
	}
}
