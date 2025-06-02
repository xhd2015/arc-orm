package field

import (
	"strings"

	"github.com/xhd2015/arc-orm/sql/expr"
)

// Expr represents a SQL condition
type Expr = expr.Expr

// Field interface represents a database field with a name and table
type Field interface {
	Name() string
	Table() string
	expr.Expr
}

// comparison represents a comparison operation between a field and a value
type comparison struct {
	field Field
	op    string
	value interface{}
}

func (c *comparison) ToSQL() (string, []interface{}, error) {
	sql, params, err := c.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " " + c.op + " ?", append(params, c.value), nil
}

// like represents a LIKE condition
type like struct {
	field Field
	value string
}

func (l *like) ToSQL() (string, []interface{}, error) {
	sql, params, err := l.field.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " LIKE ?", append(params, l.value), nil
}

// or represents an OR condition
type or struct {
	conditions []Expr
}

type and struct {
	conditions []Expr
}

// Or creates an OR condition from multiple conditions
func Or(conditions ...Expr) Expr {
	return &or{conditions: conditions}
}
func And(conditions ...Expr) Expr {
	return &and{conditions: conditions}
}

func (o *or) ToSQL() (string, []interface{}, error) {
	return joinCodnitions(o.conditions, "OR")
}

func (a *and) ToSQL() (string, []interface{}, error) {
	return joinCodnitions(a.conditions, "AND")
}

func joinCodnitions(conditions []Expr, op string) (string, []interface{}, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	sqlParts := make([]string, 0, len(conditions))
	params := make([]interface{}, 0)

	for _, cond := range conditions {
		sql, condParams, err := cond.ToSQL()
		if err != nil {
			return "", nil, err
		}
		if sql == "" {
			continue
		}
		sqlParts = append(sqlParts, sql)
		params = append(params, condParams...)
	}
	if len(sqlParts) == 0 {
		return "", nil, nil
	}
	if len(sqlParts) == 1 {
		return sqlParts[0], params, nil
	}

	return "(" + strings.Join(sqlParts, " "+op+" ") + ")", params, nil
}

// Expression represents a SQL expression
type Expression interface {
	ToExpressionSQL() (string, interface{})
}

// fieldOperation represents a field operation (like increment, decrement, concatenate)
type fieldOperation struct {
	field    Field
	operator string
	value    interface{}
}

func (op *fieldOperation) ToExpressionSQL() (string, interface{}) {
	return "`" + op.field.Table() + "`.`" + op.field.Name() + "`" + op.operator, op.value
}
