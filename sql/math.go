package sql

import "github.com/xhd2015/arc-orm/field"

func Paren(expr Expr) Expr {
	return field.Paren(expr)
}

func Add(exprs ...Expr) Expr {
	return field.Add(exprs...)
}

func Sub(exprs ...Expr) Expr {
	return field.Sub(exprs...)
}

func Mul(exprs ...Expr) Expr {
	return field.Mul(exprs...)
}

func Div(exprs ...Expr) Expr {
	return field.Div(exprs...)
}
