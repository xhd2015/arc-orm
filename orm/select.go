package orm

import (
	"context"

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
		builder: sql.Select(c.table.Fields()...).From(c.table.Name()),
		orm:     c,
	}
}

func (c *ORM[T, P]) Select(fields ...field.Field) *ORMSelectBuilder[T, P] {
	return &ORMSelectBuilder[T, P]{
		builder: sql.Select(fields...).From(c.table.Name()),
		orm:     c,
	}
}

func (c *ORMSelectBuilder[T, P]) Where(conditions ...field.Expr) *ORMSelectBuilder[T, P] {
	c.builder.Where(conditions...)
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
