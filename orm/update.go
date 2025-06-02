package orm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
	"github.com/xhd2015/less-gen/strcase"
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

// UpdateByID updates an existing record by ID with partial fields
func (o *ORM[T, P]) UpdateByID(ctx context.Context, id int64, data *P) error {
	idCondition, err := o.toIDCondition(id)
	if err != nil {
		return fmt.Errorf("failed to convert id to condition: %w", err)
	}

	return o.update(ctx, []field.Expr{idCondition}, data)
}

func (o *ORM[T, P]) UpdateBy(ctx context.Context, condition *P, data *P) error {
	if condition == nil {
		return fmt.Errorf("requires condition")
	}

	sqlConditions, err := o.ToConditions(condition)
	if err != nil {
		return fmt.Errorf("failed to convert condition to SQL conditions: %w", err)
	}

	return o.update(ctx, sqlConditions, data)
}

func (o *ORM[T, P]) update(ctx context.Context, conditions []field.Expr, data *P) error {
	if data == nil {
		return fmt.Errorf("requires data, got nil")
	}
	if len(conditions) == 0 {
		return fmt.Errorf("requires conditions")
	}

	// Create the SQL Update builder
	builder := sql.Update(o.table.Name())

	// Map struct fields to table fields
	tableFields := make(map[string]field.Field)
	for _, f := range o.table.Fields() {
		tableFields[f.Name()] = f
	}

	// Flag to track if we have any fields to update
	hasFieldsToUpdate := false

	// Check if the model has an UpdateTime field and if it's nil
	shouldAddUpdateTime := false
	hasUpdateTimeField := false
	var updateTimeField field.Field

	// Use reflection to extract non-nil fields from the partialModel
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	// Iterate through the struct fields and add them to the builder
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Special handling for UpdateTime
		if fieldType.Name == "UpdateTime" {
			hasUpdateTimeField = true
			fieldName := strcase.CamelToSnake(fieldType.Name)
			updateTimeField = tableFields[fieldName]

			// If the field is nil, we should add update_time to the query
			if field.Kind() == reflect.Ptr && field.IsNil() {
				shouldAddUpdateTime = true
			}
		}

		// Get the field value
		var fieldRValue reflect.Value
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				// Skip nil pointer fields
				continue
			}
			fieldRValue = field.Elem()
		} else {
			fieldRValue = field
		}
		fieldValue := fieldRValue.Interface()

		// Convert field name to snake_case
		fieldName := strcase.CamelToSnake(fieldType.Name)

		// Get the corresponding table field
		tableField, exists := tableFields[fieldName]
		if !exists {
			continue // Skip fields not in the table
		}

		// Convert Go value to SQL value based on type
		var sqlValue interface {
			ToExpressionSQL() (string, interface{})
		}

		switch fieldRValue.Kind() {
		case reflect.String:
			sqlValue = sql.String(fieldRValue.String())
		case reflect.Int, reflect.Int64:
			sqlValue = sql.Int64(fieldRValue.Int())
		case reflect.Int32:
			sqlValue = sql.Int32(fieldRValue.Int())
		case reflect.Float64:
			sqlValue = sql.Float64(fieldRValue.Float())
		case reflect.Bool:
			sqlValue = sql.Bool(fieldRValue.Bool())
		case reflect.Struct:
			// Handle time.Time specially
			if t, ok := fieldValue.(time.Time); ok {
				sqlValue = sql.Time(t)
			}
		}

		// Skip if we couldn't convert the value
		if sqlValue == nil {
			continue
		}

		// Add to the builder
		builder.Set(tableField, sqlValue)
		hasFieldsToUpdate = true
	}

	// Check if there are any fields to update
	if !hasFieldsToUpdate {
		return ErrNothingToUpdate
	}

	// If we have an UpdateTime field that was nil, add it to the query with current time
	if hasUpdateTimeField && shouldAddUpdateTime {
		builder.Set(updateTimeField, sql.Time(time.Now()))
	}

	// Add WHERE clause for ID
	builder.Where(conditions...)

	// Generate the SQL and args
	query, args, err := builder.SQL()
	if err != nil {
		return fmt.Errorf("failed to build update SQL: %w", err)
	}

	// Execute the update
	err = o.engine.GetEngine().Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to execute UpdateByID: %w", err)
	}

	return nil
}

func (c *ORMUpdateBuilder[T, P]) Set(f field.Field, value field.Expression) *ORMUpdateBuilder[T, P] {
	c.builder.Set(f, value)
	return c
}

func (c *ORMUpdateBuilder[T, P]) Where(conditions ...field.Expr) *ORMUpdateBuilder[T, P] {
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
