package orm

import (
	"context"
	"fmt"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
	"github.com/xhd2015/arc-orm/sql/expr"
)

type ORMSelectBuilder[T any, P any] struct {
	builder *sql.SelectBuilder
	orm     *ORM[T, P]
}

func (c *ORM[T, P]) SelectAll() *ORMSelectBuilder[T, P] {
	return &ORMSelectBuilder[T, P]{
		builder: sql.Select(fieldsToExprs(c.table.Fields())...).From(c.table.Name()),
		orm:     c,
	}
}

func (c *ORM[T, P]) Select(fields ...field.Field) *ORMSelectBuilder[T, P] {
	return &ORMSelectBuilder[T, P]{
		builder: sql.Select(fieldsToExprs(fields)...).From(c.table.Name()),
		orm:     c,
	}
}

// SelectExpr creates a select query with arbitrary expressions.
// Use this for queries with aggregations, functions, or custom expressions
// that don't map directly to entity fields.
// Example:
//
//	orm.SelectExpr(sql.Date(field).As("date"), sql.Count(sql.All).As("count"))
func (c *ORM[T, P]) SelectExpr(exprs ...sql.Expr) *ORMSelectBuilder[T, P] {
	return &ORMSelectBuilder[T, P]{
		builder: sql.Select(exprs...).From(c.table.Name()),
		orm:     c,
	}
}

func fieldsToExprs(fields []field.Field) []sql.Expr {
	exprs := make([]sql.Expr, 0, len(fields))
	for _, field := range fields {
		exprs = append(exprs, field)
	}
	return exprs
}

func (c *ORMSelectBuilder[T, P]) Exclude(fields ...field.Field) *ORMSelectBuilder[T, P] {
	c.builder.Exclude(fields...)
	return c
}

func (c *ORMSelectBuilder[T, P]) Where(conditions ...field.Expr) *ORMSelectBuilder[T, P] {
	c.builder.Where(conditions...)
	return c
}

func (c *ORMSelectBuilder[T, P]) LeftJoin(tableName string, condition field.Expr) *ORMSelectBuilder[T, P] {
	c.builder.LeftJoin(tableName, condition)
	return c
}

func (c *ORMSelectBuilder[T, P]) GroupBy(exprs ...expr.Expr) *ORMSelectBuilder[T, P] {
	c.builder.GroupBy(exprs...)
	return c
}

func (c *ORMSelectBuilder[T, P]) Having(conditions ...field.Expr) *ORMSelectBuilder[T, P] {
	c.builder.Having(conditions...)
	return c
}

func (c *ORMSelectBuilder[T, P]) OrderBy(orderFields ...expr.Expr) *ORMSelectBuilder[T, P] {
	c.builder.OrderBy(orderFields...)
	return c
}

func (c *ORMSelectBuilder[T, P]) Limit(limit int) *ORMSelectBuilder[T, P] {
	c.builder.Limit(limit)
	return c
}

func (c *ORMSelectBuilder[T, P]) Offset(offset int) *ORMSelectBuilder[T, P] {
	c.builder.Offset(offset)
	return c
}

func (c *ORMSelectBuilder[T, P]) Query(ctx context.Context) ([]*T, error) {
	sql, args, err := c.builder.SQL()
	if err != nil {
		return nil, err
	}
	return c.orm.QuerySQL(ctx, sql, args)
}

func (c *ORMSelectBuilder[T, P]) QueryOne(ctx context.Context) (*T, error) {
	c.builder.Limit(1)
	sql, args, err := c.builder.SQL()
	if err != nil {
		return nil, err
	}
	list, err := c.orm.QuerySQL(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (c *ORMSelectBuilder[T, P]) RequireOne(ctx context.Context) (*T, error) {
	result, err := c.QueryOne(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("record not found")
	}
	return result, nil
}

// QueryInto executes the query and scans results into the provided slice pointer.
// This is useful for queries with custom SELECT expressions (like aggregations)
// that don't match the ORM's entity type.
// Example:
//
//	type Result struct {
//	    Date  string `json:"date"`
//	    Count int    `json:"count"`
//	}
//	var results []*Result
//	err := orm.SelectExpr(sql.Date(field), sql.Count(sql.All).As("count")).
//	    Where(...).GroupBy(sql.Date(field)).QueryInto(ctx, &results)
func (c *ORMSelectBuilder[T, P]) QueryInto(ctx context.Context, result interface{}) error {
	sqlStr, args, err := c.builder.SQL()
	if err != nil {
		return err
	}
	return c.orm.engine.GetEngine().Query(ctx, sqlStr, args, result)
}
