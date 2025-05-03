package sql

import "github.com/xhd2015/arc-orm/field"

// Or creates an OR condition from multiple conditions
func Or(conditions ...field.Condition) field.Condition {
	return field.Or(conditions...)
}

func And(conditions ...field.Condition) field.Condition {
	return field.And(conditions...)
}
