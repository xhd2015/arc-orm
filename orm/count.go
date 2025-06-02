package orm

import (
	"context"
	"fmt"
	"reflect"
)

// Count executes a count query and returns the matching records
// The model must have a Count field of type int64 to receive the count value
func (o *ORM[T, P]) Count(ctx context.Context, query string, args []interface{}) ([]*T, error) {
	// TODO: make this validate once when creating the ORM instance
	// Validate that type T has a Count field of type int64
	modelType := reflect.TypeOf((*T)(nil)).Elem()

	// Find the Count field
	countField, found := modelType.FieldByName("Count")
	if !found {
		return nil, ErrMissingCountField
	}

	// Validate the Count field type is int64
	int64Type := reflect.TypeOf(int64(0))
	if countField.Type != int64Type {
		return nil, fmt.Errorf("%w, got %s", ErrMissingCountField, countField.Type.String())
	}

	// Execute the query using the Query method
	return o.QuerySQL(ctx, query, args)
}
