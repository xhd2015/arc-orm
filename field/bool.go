package field

// BoolField represents a boolean database field
// In MySQL, boolean values are typically stored as TINYINT(1) where 0 = false and 1 = true
type BoolField struct {
	FieldName string
	TableName string
}

// Name returns the field name
func (f BoolField) Name() string {
	return f.FieldName
}

// Table returns the table name
func (f BoolField) Table() string {
	return f.TableName
}

// ToSQL returns the SQL representation of the field
func (f BoolField) ToSQL() (string, []interface{}, error) {
	if f.TableName == "" {
		return "`" + f.FieldName + "`", nil, nil
	}
	return "`" + f.TableName + "`.`" + f.FieldName + "`", nil, nil
}

// Eq creates an equality condition (field = value)
// The boolean value is converted to 1 (true) or 0 (false) for SQL
func (f BoolField) Eq(value bool) Expr {
	sqlValue := int32(0)
	if value {
		sqlValue = 1
	}
	return &comparison{
		field: f,
		op:    "=",
		value: sqlValue,
	}
}

// IsTrue creates a condition checking if the field is true (field = 1)
func (f BoolField) IsTrue() Expr {
	return f.Eq(true)
}

// IsFalse creates a condition checking if the field is false (field = 0)
func (f BoolField) IsFalse() Expr {
	return f.Eq(false)
}

// IsNull creates an IS NULL condition (field IS NULL)
func (f BoolField) IsNull() Expr {
	return &nullCondition{
		field:  f,
		isNull: true,
	}
}

// IsNotNull creates an IS NOT NULL condition (field IS NOT NULL)
func (f BoolField) IsNotNull() Expr {
	return &nullCondition{
		field:  f,
		isNull: false,
	}
}

// Asc returns an ascending order specification for this field
func (f BoolField) Asc() OrderField {
	return OrderField{field: f, desc: false}
}

// Desc returns a descending order specification for this field
func (f BoolField) Desc() OrderField {
	return OrderField{field: f, desc: true}
}

// As returns this field with an alias
func (f BoolField) As(alias string) Field {
	return As(f, alias)
}
