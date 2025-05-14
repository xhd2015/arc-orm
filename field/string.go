package field

// StringField represents a string database field
type StringField struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f StringField) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f StringField) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f StringField) ToSQL() string {
	return "`" + f.TableName + "`.`" + f.FieldName + "`"
}

// Eq creates an equality condition (field = value)
func (f StringField) Eq(value string) Condition {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// Lte creates a less than or equal to condition (field <= value)
func (f StringField) Lte(value string) Condition {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

// Lt creates a less than condition (field < value)
func (f StringField) Lt(value string) Condition {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

// Gte creates a greater than or equal to condition (field >= value)
func (f StringField) Gte(value string) Condition {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

// Gt creates a greater than condition (field > value)
func (f StringField) Gt(value string) Condition {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

type noOp struct {
}

func (f noOp) ToSQL() (string, []interface{}, error) {
	return "", nil, nil
}

func (f StringField) In(values ...string) Condition {
	if len(values) == 0 {
		panic("in requires non-empty values")
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

func (f StringField) InOrEmpty(values ...string) Condition {
	if len(values) == 0 {
		return noOp{}
	}
	return f.In(values...)
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f StringField) EqField(other Field) Condition {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f StringField) Neq(value string) Condition {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

// Like creates a LIKE condition (field LIKE value)
func (f StringField) Like(value string) Condition {
	return &like{
		field: f,
		value: value,
	}
}

// Contains creates a LIKE condition with wildcards (field LIKE %value%)
func (f StringField) Contains(value string) Condition {
	if value == "" {
		return noOp{}
	}
	return &like{
		field: f,
		value: "%" + value + "%",
	}
}

// StartsWith creates a LIKE condition with wildcard (field LIKE value%)
func (f StringField) StartsWith(value string) Condition {
	if value == "" {
		return noOp{}
	}
	return &like{
		field: f,
		value: value + "%",
	}
}

// EndsWith creates a LIKE condition with wildcard (field LIKE %value)
func (f StringField) EndsWith(value string) Condition {
	if value == "" {
		return noOp{}
	}
	return &like{
		field: f,
		value: "%" + value,
	}
}

// Asc returns an ascending order specification for this field
func (f StringField) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f StringField) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f StringField) As(alias string) Field {
	return As(f, alias)
}

// Concat returns an expression to concatenate this field with a string
func (f StringField) Concat(value string) Expression {
	return &fieldOperation{
		field:    f,
		operator: "+",
		value:    value,
	}
}
