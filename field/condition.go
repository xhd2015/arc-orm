package field

import (
	"strings"

	"github.com/xhd2015/arc-orm/sql/expr"
)

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
	sql, params, err := c.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " IN (" + strings.Join(placeholders, ", ") + ")", append(params, c.values...), nil
}

// OrderField represents a field with ordering direction
type OrderField struct {
	field Field
	desc  bool
}

// ToSQL returns the SQL for ordering
func (o OrderField) ToSQL() (string, []interface{}, error) {
	sql, params, err := o.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	if o.desc {
		sql += " DESC"
	} else {
		sql += " ASC"
	}
	return sql, params, nil
}

// fieldComparison represents a comparison between two fields
type fieldComparison struct {
	left  expr.Expr
	op    string
	right expr.Expr
}

func (c *fieldComparison) ToSQL() (string, []interface{}, error) {
	leftSQL, leftParams, err := c.left.ToSQL()
	if err != nil {
		return "", nil, err
	}
	rightSQL, rightParams, err := c.right.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return leftSQL + " " + c.op + " " + rightSQL, concatParams(leftParams, rightParams), nil
}

// nullCondition represents an IS NULL or IS NOT NULL condition
type nullCondition struct {
	field  Field
	isNull bool
}

func (c *nullCondition) ToSQL() (string, []interface{}, error) {
	sql, params, err := c.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	if c.isNull {
		return sql + " IS NULL", params, nil
	}
	return sql + " IS NOT NULL", params, nil
}

// between represents a BETWEEN condition
type between struct {
	field Field
	start interface{}
	end   interface{}
}

func (b *between) ToSQL() (string, []interface{}, error) {
	sql, params, err := b.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " BETWEEN ? AND ?", concatParams(params, []interface{}{b.start, b.end}), nil
}

type betweenExpr struct {
	field Field
	start expr.Expr
	end   expr.Expr
}

func (b *betweenExpr) ToSQL() (string, []interface{}, error) {
	sql, params, err := b.field.ToSQL()
	if err != nil {
		return "", nil, err
	}

	startSQL, startParams, err := b.start.ToSQL()
	if err != nil {
		return "", nil, err
	}
	endSQL, endParams, err := b.end.ToSQL()
	if err != nil {
		return "", nil, err
	}

	return sql + " BETWEEN " + startSQL + " AND " + endSQL, concatParams(params, startParams, endParams), nil
}

func concatParams(params ...[]interface{}) []interface{} {
	n := 0
	for _, p := range params {
		n += len(p)
	}
	result := make([]interface{}, n)
	offset := 0
	for _, p := range params {
		copy(result[offset:], p)
		offset += len(p)
	}
	return result
}
