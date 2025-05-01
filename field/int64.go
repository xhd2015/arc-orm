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

// inCondition represents an IN condition
type inCondition struct {
	field  Field
	values []interface{}
}

func (c *inCondition) ToSQL() (string, []interface{}, error) {
	if len(c.values) == 0 {
		return "1=0", nil, nil // Empty IN clause should always be false
	}

	placeholders := make([]string, len(c.values))
	for i := range c.values {
		placeholders[i] = "?"
	}

	return c.field.ToSQL() + " IN (" + joinStrings(placeholders, ", ") + ")", c.values, nil
}

// OrderField represents a field with ordering direction
type OrderField struct {
	field Field
	desc  bool
}

// ToSQL returns the SQL for ordering
func (o OrderField) ToSQL() string {
	sql := o.field.ToSQL()
	if o.desc {
		sql += " DESC"
	} else {
		sql += " ASC"
	}
	return sql
}

// fieldComparison represents a comparison between two fields
type fieldComparison struct {
	left  Field
	op    string
	right Field
}

func (c *fieldComparison) ToSQL() (string, []interface{}, error) {
	return c.left.ToSQL() + " " + c.op + " " + c.right.ToSQL(), nil, nil
}
