package sql

import "github.com/xhd2015/arc-orm/field"

// Or creates an OR condition from multiple conditions
func Or(conditions ...field.Expr) field.Expr {
	return field.Or(conditions...)
}

func And(conditions ...field.Expr) field.Expr {
	return field.And(conditions...)
}
