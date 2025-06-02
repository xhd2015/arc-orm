package field

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
func (f TimeField) ToSQL() (string, []interface{}, error) {
	if f.TableName == "" {
		return "`" + f.FieldName + "`", nil, nil
	}
	return "`" + f.TableName + "`.`" + f.FieldName + "`", nil, nil
}

// Eq creates an equality condition (field = value)
func (f TimeField) Eq(value string) Expr {
	return &comparison{
		field: f,
		op:    "=",
		value: value,
	}
}

// EqField creates an equality condition between two fields (field1 = field2)
func (f TimeField) EqField(other Field) Expr {
	return &fieldComparison{
		left:  f,
		op:    "=",
		right: other,
	}
}

// Neq creates a not equal condition (field != value)
func (f TimeField) Neq(value string) Expr {
	return &comparison{
		field: f,
		op:    "!=",
		value: value,
	}
}

func (f TimeField) NeqField(other TimeField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "!=",
		right: other,
	}
}

// Gt creates a greater than condition (field > value)
func (f TimeField) Gt(value string) Expr {
	return &comparison{
		field: f,
		op:    ">",
		value: value,
	}
}

func (f TimeField) GtField(other TimeField) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">",
		right: other,
	}
}

// Gte creates a greater than or equal condition (field >= value)
func (f TimeField) Gte(value string) Expr {
	return &comparison{
		field: f,
		op:    ">=",
		value: value,
	}
}

func (f TimeField) GteField(other TimeField) Expr {
	return &fieldComparison{
		left:  f,
		op:    ">=",
		right: other,
	}
}

// Lt creates a less than condition (field < value)
func (f TimeField) Lt(value string) Expr {
	return &comparison{
		field: f,
		op:    "<",
		value: value,
	}
}

func (f TimeField) LtField(other TimeField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<",
		right: other,
	}
}

// Lte creates a less than or equal condition (field <= value)
func (f TimeField) Lte(value string) Expr {
	return &comparison{
		field: f,
		op:    "<=",
		value: value,
	}
}

func (f TimeField) LteField(other TimeField) Expr {
	return &fieldComparison{
		left:  f,
		op:    "<=",
		right: other,
	}
}

// Between creates a BETWEEN condition
func (f TimeField) Between(start string, end string) Expr {
	return &between{
		field: f,
		start: start,
		end:   end,
	}
}

func (f TimeField) BetweenField(start TimeField, end TimeField) Expr {
	return &betweenExpr{
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
