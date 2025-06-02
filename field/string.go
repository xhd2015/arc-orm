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
func (f StringField) ToSQL() (string, []interface{}, error) {
	return "`" + f.TableName + "`.`" + f.FieldName + "`", nil, nil
}

// Eq creates an equality condition (field = value)
func (f StringField) Eq(value string) Expr {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f StringField) EqField(other Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f StringField) Neq(value string) Expr {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

func (f StringField) NeqField(other StringField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "!=",
		right: other,
	}
}

// Lte creates a less than or equal to condition (field <= value)
func (f StringField) Lte(value string) Expr {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

func (f StringField) LteField(other StringField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<=",
		right: other,
	}
}

// Lt creates a less than condition (field < value)
func (f StringField) Lt(value string) Expr {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

func (f StringField) LtField(other StringField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<",
		right: other,
	}
}

// Gte creates a greater than or equal to condition (field >= value)
func (f StringField) Gte(value string) Expr {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

func (f StringField) GteField(other StringField) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">=",
		right: other,
	}
}

// Gt creates a greater than condition (field > value)
func (f StringField) Gt(value string) Expr {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

func (f StringField) GtField(other StringField) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">",
		right: other,
	}
}

type noOp struct {
}

func (f noOp) ToSQL() (string, []interface{}, error) {
	return "", nil, nil
}

func (f StringField) In(values ...string) Expr {
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

func (f StringField) InOrEmpty(values ...string) Expr {
	if len(values) == 0 {
		return noOp{}
	}
	return f.In(values...)
}

// Like creates a LIKE condition (field LIKE value)
func (f StringField) Like(value string) Expr {
	return &like{
		field: f,
		value: value,
	}
}

// Contains creates a LIKE condition with wildcards (field LIKE %value%)
func (f StringField) Contains(value string) Expr {
	if value == "" {
		return noOp{}
	}
	return &like{
		field: f,
		value: "%" + value + "%",
	}
}

// StartsWith creates a LIKE condition with wildcard (field LIKE value%)
func (f StringField) StartsWith(value string) Expr {
	if value == "" {
		return noOp{}
	}
	return &like{
		field: f,
		value: value + "%",
	}
}

// EndsWith creates a LIKE condition with wildcard (field LIKE %value)
func (f StringField) EndsWith(value string) Expr {
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
