package orm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
	"github.com/xhd2015/less-gen/strcase"
)

// Insert adds a new record to the database and returns the generated ID
func (o *ORM[T, P]) Insert(ctx context.Context, model *T) (int64, error) {
	// Use reflection to extract field values from the model
	if model == nil {
		return 0, errors.New("model cannot be nil")
	}

	// Get the reflect.Value of the model struct (dereference the pointer)
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	// Create the SQL Insert builder
	builder := sql.InsertInto(o.table.Name())

	// Map struct fields to table fields
	tableFields := make(map[string]field.Field)
	for _, f := range o.table.Fields() {
		tableFields[f.Name()] = f
	}

	// Iterate through the struct fields and add them to the builder
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Skip Count field (if present)
		if fieldType.Name == "Count" {
			continue
		}

		// Convert field name to snake_case
		fieldName := strcase.CamelToSnake(fieldType.Name)

		// Get the corresponding table field
		tableField, exists := tableFields[fieldName]
		if !exists {
			return 0, fmt.Errorf("field %s not found in table %s", fieldName, o.table.Name())
		}

		// Convert Go value to SQL value based on type
		var sqlValue interface {
			ToExpressionSQL() (string, interface{})
		}
		switch field.Kind() {
		case reflect.String:
			sqlValue = sql.String(field.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sqlValue = sql.Int64(field.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sqlValue = sql.Int64(field.Uint())
		case reflect.Float64, reflect.Float32:
			sqlValue = sql.Float64(field.Float())
		case reflect.Bool:
			sqlValue = sql.Bool(field.Bool())
		case reflect.Struct:
			// Handle time.Time specially
			if fieldType.Type.String() == "time.Time" {
				timeValue := field.Interface().(time.Time)

				// Auto-fill CreateTime and UpdateTime with current time if they're zero
				if (fieldType.Name == "CreateTime" || fieldType.Name == "UpdateTime") && timeValue.IsZero() {
					timeValue = time.Now()
				}

				sqlValue = sql.Time(timeValue)
			}
		}

		// Skip if we couldn't convert the value
		if sqlValue == nil {
			return 0, fmt.Errorf("failed to convert field %s to SQL value: %s", fieldType.Name, field.Type())
		}

		// Add to the builder
		builder.Set(tableField, sqlValue)
	}

	// Generate the SQL and args
	query, args, err := builder.SQL()
	if err != nil {
		return 0, fmt.Errorf("failed to build insert SQL: %w", err)
	}

	// Execute the insert and get the ID
	id, err := o.engine.GetEngine().ExecInsert(ctx, query, args)
	if err != nil {
		return 0, fmt.Errorf("failed to execute Insert: %w", err)
	}

	return id, nil
}
