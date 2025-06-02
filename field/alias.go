package field

// AliasField wraps any field with an alias
type AliasField struct {
	field Field
	alias string
}

// As creates an aliased field
func As(f Field, alias string) *AliasField {
	return &AliasField{
		field: f,
		alias: alias,
	}
}

// Name returns the field name
func (a *AliasField) Name() string {
	return a.field.Name()
}

// Table returns the table name
func (a *AliasField) Table() string {
	return a.field.Table()
}

// ToSQL returns the SQL representation of the field with its alias
func (a *AliasField) ToSQL() (string, []interface{}, error) {
	sql, params, err := a.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " AS `" + a.alias + "`", params, nil
}
