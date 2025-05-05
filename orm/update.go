package orm

import (
	"context"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
)

type ORMUpdateBuilder[T any, P any] struct {
	builder *sql.UpdateBuilder
	orm     *ORM[T, P]
}

func (c *ORM[T, P]) Update() *ORMUpdateBuilder[T, P] {
	return &ORMUpdateBuilder[T, P]{
		builder: sql.Update(c.table.Name()),
		orm:     c,
	}
}

func (c *ORMUpdateBuilder[T, P]) Set(f field.Field, value field.Expression) *ORMUpdateBuilder[T, P] {
	c.builder.Set(f, value)
	return c
}

func (c *ORMUpdateBuilder[T, P]) Where(conditions ...field.Condition) *ORMUpdateBuilder[T, P] {
	c.builder.Where(conditions...)
	return c
}

func (c *ORMUpdateBuilder[T, P]) Exec(ctx context.Context) error {
	sql, args, err := c.builder.SQL()
	if err != nil {
		return err
	}
	return c.orm.engine.GetEngine().Exec(ctx, sql, args)
}
