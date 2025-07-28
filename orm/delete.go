package orm

import (
	"context"
	"fmt"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
)

// DeleteByID deletes a record by its ID
func (o *ORM[T, P]) DeleteByID(ctx context.Context, id int64) error {
	idCondition, err := o.toIDCondition(id)
	if err != nil {
		return fmt.Errorf("failed to convert id to condition: %w", err)
	}

	return o.deleteBy(ctx, []field.Expr{idCondition})
}

// DeleteByID deletes a record by its ID
func (o *ORM[T, P]) DeleteBy(ctx context.Context, condition *P) error {
	if condition == nil {
		return fmt.Errorf("requires condition")
	}

	sqlConditions, err := o.ToConditions(condition)
	if err != nil {
		return fmt.Errorf("failed to convert condition to SQL conditions: %w", err)
	}

	return o.deleteBy(ctx, sqlConditions)
}

// DeleteByID deletes a record by its ID
func (o *ORM[T, P]) DeleteWhere(ctx context.Context, conditions ...field.Expr) error {
	if len(conditions) == 0 {
		return fmt.Errorf("requires conditions")
	}

	return o.deleteBy(ctx, conditions)
}

func (o *ORM[T, P]) deleteBy(ctx context.Context, conditions []field.Expr) error {
	if len(conditions) == 0 {
		return fmt.Errorf("requires conditions")
	}

	// Create the SQL Delete builder
	query, args, err := sql.DeleteFrom(o.table.Name()).
		Where(conditions...).
		SQL()

	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	// Execute the delete
	err = o.engine.GetEngine().Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to execute DeleteByID: %w", err)
	}

	return nil
}
