package orm

import (
	"context"
	"errors"
	"fmt"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
)

// QuerySQL executes the provided SQL query and returns matching records
func (o *ORM[T, P]) QuerySQL(ctx context.Context, sql string, args []interface{}) ([]*T, error) {
	// Create a slice to hold the results
	var results []*T

	// Execute the query using the engine
	err := o.engine.GetEngine().Query(ctx, sql, args, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, nil
}

// GetByID retrieves a record by its primary key
// the record must exist, otherwise it will return an error
func (o *ORM[T, P]) GetByID(ctx context.Context, id int64) (*T, error) {
	idCondition, err := o.toIDCondition(id)
	if err != nil {
		return nil, fmt.Errorf("failed to convert id to condition: %w", err)
	}
	return o.get(ctx, []field.Expr{idCondition})
}

func (o *ORM[T, P]) GetBy(ctx context.Context, condition *P) (*T, error) {
	if condition == nil {
		return nil, fmt.Errorf("requires condition")
	}

	sqlConditions, err := o.ToConditions(condition)
	if err != nil {
		return nil, fmt.Errorf("failed to convert condition to SQL conditions: %w", err)
	}

	return o.get(ctx, sqlConditions)
}

func (o *ORM[T, P]) get(ctx context.Context, conditions []field.Expr) (*T, error) {
	querySQL, args, err := sql.Select(fieldsToExprs(o.table.Fields())...).
		From(o.table.Name()).
		Where(conditions...).
		Limit(1).
		SQL()
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}

	// Create a slice to hold the result
	var results []*T

	// Execute the query
	err = o.engine.GetEngine().Query(ctx, querySQL, args, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Get: %w", err)
	}

	// Check if we found a result
	if len(results) == 0 {
		return nil, errors.New("data not found")
	}

	return results[0], nil
}
