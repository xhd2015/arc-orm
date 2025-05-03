package field

import "strings"

// inCondition represents an IN condition
type inCondition struct {
	field  Field
	values []interface{}
}

func (c *inCondition) ToSQL() (string, []interface{}, error) {
	n := len(c.values)
	if n == 0 {
		return "", nil, nil
	}
	placeholders := make([]string, n)
	for i := 0; i < n; i++ {
		placeholders[i] = "?"
	}

	return c.field.ToSQL() + " IN (" + strings.Join(placeholders, ", ") + ")", c.values, nil
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

// nullCondition represents an IS NULL or IS NOT NULL condition
type nullCondition struct {
	field  Field
	isNull bool
}

func (c *nullCondition) ToSQL() (string, []interface{}, error) {
	if c.isNull {
		return c.field.ToSQL() + " IS NULL", nil, nil
	}
	return c.field.ToSQL() + " IS NOT NULL", nil, nil
}

// between represents a BETWEEN condition
type between struct {
	field Field
	start interface{}
	end   interface{}
}

func (b *between) ToSQL() (string, []interface{}, error) {
	return b.field.ToSQL() + " BETWEEN ? AND ?", []interface{}{b.start, b.end}, nil
}
