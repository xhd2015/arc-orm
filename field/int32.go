package field

import "fmt"

// Int32Field represents an int64 database field
type Int32Field struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f Int32Field) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f Int32Field) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f Int32Field) ToSQL() (string, []interface{}, error) {
	// If the field has no table, or the table name is empty, just use the field name
	if f.TableName == "" {
		return "`" + f.FieldName + "`", nil, nil
	}
	return "`" + f.TableName + "`.`" + f.FieldName + "`", nil, nil
}

// Eq creates an equality condition (field = value)
func (f Int32Field) Eq(value int32) Expr {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f Int32Field) EqField(other Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f Int32Field) Neq(value int32) Expr {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

func (f Int32Field) NeqField(other Int32Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "!=",
		right: other,
	}
}

// Gt creates a greater than condition (field > value)
func (f Int32Field) Gt(value int32) Expr {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

func (f Int32Field) GtField(other Int32Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">",
		right: other,
	}
}

// Gte creates a greater than or equal condition (field >= value)
func (f Int32Field) Gte(value int32) Expr {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

func (f Int32Field) GteField(other Int32Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">=",
		right: other,
	}
}

// Lt creates a less than condition (field < value)
func (f Int32Field) Lt(value int32) Expr {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

func (f Int32Field) LtField(other Int32Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<",
		right: other,
	}
}

// Lte creates a less than or equal condition (field <= value)
func (f Int32Field) Lte(value int32) Expr {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

func (f Int32Field) LteField(other Int32Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<=",
		right: other,
	}
}

// IsNull creates an IS NULL condition (field IS NULL)
func (f Int32Field) IsNull() Expr {
	return &nullCondition{
		field:  f,
		isNull: true,
	}
}

func (f Int32Field) IsNotNull() Expr {
	return &nullCondition{
		field:  f,
		isNull: false,
	}
}

// In creates an IN condition (field IN (values))
func (f Int32Field) In(values ...int32) Expr {
	if len(values) == 0 {
		panic(fmt.Errorf("in requires at least one value"))
	}
	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}
	return &inCondition{
		field:  f,
		values: interfaceValues,
	}
}

// InOrEmpty creates an IN condition (field IN (values))
func (f Int32Field) InOrEmpty(values ...int32) Expr {
	if len(values) == 0 {
		return noOp{}
	}
	return f.In(values...)
}

// Asc returns an ascending order specification for this field
func (f Int32Field) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f Int32Field) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f Int32Field) As(alias string) Field {
	return As(f, alias)
}

// Increment returns an expression to increment this field by a value
func (f Int32Field) Increment(value int32) Expr {
	return &fieldOperation{
		field:    f,
		operator: "+",
		value:    value,
	}
}

// Decrement returns an expression to decrement this field by a value
func (f Int32Field) Decrement(value int32) Expr {
	return &fieldOperation{
		field:    f,
		operator: "-",
		value:    value,
	}
}
