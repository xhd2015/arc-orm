package orm

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
)

type ORMCountBuilder[T any, P any] struct {
	builder *sql.SelectBuilder
	orm     *ORM[T, P]
}

// Count executes a count query and returns the matching records
// The model must have a Count field of type int64 to receive the count value
func (c *ORM[T, P]) Count(fields ...sql.Expr) *ORMCountBuilder[T, P] {
	// TODO: make this validate once when creating the ORM instance
	// Validate that type T has a Count field of type int64
	modelType := reflect.TypeOf((*T)(nil)).Elem()

	// Find the Count field
	_, found := modelType.FieldByName("Count")
	if !found {
		panic(ErrMissingCountField)
	}

	allFields := make([]sql.Expr, 0, len(fields)+1)
	countFieldExpr := sql.Count(sql.All).As("count")

	allFields = append(allFields, countFieldExpr)
	allFields = append(allFields, fields...)

	return &ORMCountBuilder[T, P]{
		builder: sql.Select(allFields...).From(c.table.Name()),
		orm:     c,
	}
}

func (c *ORMCountBuilder[T, P]) Exclude(fields ...field.Field) *ORMCountBuilder[T, P] {
	c.builder.Exclude(fields...)
	return c
}

func (c *ORMCountBuilder[T, P]) Where(conditions ...field.Expr) *ORMCountBuilder[T, P] {
	c.builder.Where(conditions...)
	return c
}

func (c *ORMCountBuilder[T, P]) GroupBy(fields ...field.Field) *ORMCountBuilder[T, P] {
	c.builder.GroupBy(fields...)
	return c
}

func (c *ORMCountBuilder[T, P]) Limit(limit int) *ORMCountBuilder[T, P] {
	c.builder.Limit(limit)
	return c
}

func (c *ORMCountBuilder[T, P]) Offset(offset int) *ORMCountBuilder[T, P] {
	c.builder.Offset(offset)
	return c
}

func (c *ORMCountBuilder[T, P]) Query(ctx context.Context) (int64, error) {
	one, err := c.QueryOneData(ctx)
	if err != nil {
		return 0, err
	}
	if one == nil {
		return 0, fmt.Errorf("count query expect at least one row")
	}
	count := reflect.ValueOf(one).Elem().FieldByName("Count").Int()
	return int64(count), nil
}

func (c *ORMCountBuilder[T, P]) QueryMany(ctx context.Context) ([]*T, error) {
	sql, args, err := c.builder.SQL()
	if err != nil {
		return nil, err
	}
	return c.orm.QuerySQL(ctx, sql, args)
}

func (c *ORMCountBuilder[T, P]) QueryOneData(ctx context.Context) (*T, error) {
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
