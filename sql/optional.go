package sql

import "github.com/xhd2015/arc-orm/field"

// Optional conditionally includes conditions based on the boolean flag v.
// If v is true, all conditions in conds are included.
// If v is false, returns an empty condition that produces no SQL.
func Optional(v bool, conds ...field.Expr) field.Expr {
	if !v {
		return noOp{}
	}
	if len(conds) == 0 {
		return noOp{}
	}
	if len(conds) == 1 {
		return conds[0]
	}
	return field.And(conds...)
}

// noOp represents a condition that produces no SQL
type noOp struct{}

// ToSQL returns empty SQL for the empty condition
func (c noOp) ToSQL() (string, []interface{}, error) {
	return "", nil, nil
}
