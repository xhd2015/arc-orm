package sql

import (
	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql/expr"
)

// Or creates an OR condition from multiple conditions
func Or(conditions ...expr.Expr) expr.Expr {
	return field.Or(conditions...)
}

func And(conditions ...expr.Expr) expr.Expr {
	return field.And(conditions...)
}

type not struct {
	conditions []Expr
}

func Not(conditions ...expr.Expr) *not {
	return &not{
		conditions: conditions,
	}
}

func (n *not) ToSQL() (string, []interface{}, error) {
	if len(n.conditions) == 0 {
		return "", nil, nil
	}

	var cond Expr
	if len(n.conditions) > 1 {
		cond = And(n.conditions...)
	} else {
		cond = n.conditions[0]
	}

	condSQL, condParams, err := cond.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return "NOT (" + condSQL + ")", condParams, nil
}
