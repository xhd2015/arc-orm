package sql

import (
	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql/expr"
)

type Expr = expr.Expr

var All field.Field = stringField("*")
var STAR = All

type stringField string

func (s stringField) Name() string {
	return "*"
}

func (s stringField) Table() string {
	return ""
}

func (s stringField) ToSQL() (string, []interface{}, error) {
	return "*", nil, nil
}
