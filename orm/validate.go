package orm

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/table"
	"github.com/xhd2015/less-gen/strcase"
)

// Errors returned by validation
var (
	ErrNotStruct          = errors.New("model type must be a struct")
	ErrNotPointerStruct   = errors.New("optional fields type must be a struct with pointer fields")
	ErrFieldMismatch      = errors.New("field mismatch between model and table")
	ErrFieldTypeMismatch  = errors.New("field type mismatch between model and table")
	ErrFieldCountMismatch = errors.New("number of fields in model does not match table")
	ErrInvalidFieldNaming = errors.New("field name must be strict CamelCase (no consecutive uppercase letters)")
)

// Validate checks if the model type T and optional fields type P
// match the table definition.
func (o *ORM[T, P]) Validate() error {
	// Validate model type
	if err := validateModelType[T](o.table); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	// Validate optional fields type
	if err := validateOptionalType[T, P](); err != nil {
		return fmt.Errorf("optional fields validation failed: %w", err)
	}

	return nil
}

// validateModelType checks if the model type T is a struct and its fields
// match the table definition.
func validateModelType[T any](tbl table.Table) error {
	// Get the reflect.Type of T
	modelType := reflect.TypeOf((*T)(nil)).Elem()

	// Check if T is a struct
	if modelType.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// Get table fields
	tableFields := tbl.Fields()

	// Check if table has a 'count' field - this is not allowed
	for _, f := range tableFields {
		if f.Name() == "count" {
			return fmt.Errorf("table must not contain a 'count' field, it is reserved for query operations")
		}
	}

	// Build maps for field comparison - use snake_case for keys
	tableFieldMap := make(map[string]field.Field)
	for _, f := range tableFields {
		// Check for create_time and update_time in table fields
		if f.Name() == "create_time" || f.Name() == "update_time" {
			// Ensure they are TimeField type
			if _, ok := f.(field.TimeField); !ok {
				return fmt.Errorf("table field '%s' must be of type TimeField", f.Name())
			}
		}
		tableFieldMap[f.Name()] = f
	}

	modelFieldMap := make(map[string]reflect.StructField)
	modelHasCountField := false
	countFieldType := reflect.TypeOf(int64(0))
	var countField reflect.StructField
	timeType := reflect.TypeOf(time.Time{})

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.IsExported() {
			// Validate field naming - must be strict CamelCase (no consecutive uppercase)
			if err := validateFieldNaming(field.Name); err != nil {
				return err
			}

			fieldName := getFieldName(field)

			// Special handling for Count field
			if field.Name == "Count" {
				modelHasCountField = true
				countField = field

				// Skip adding it to the map to avoid comparison with table fields
				continue
			}

			// Check for CreateTime and UpdateTime fields
			if field.Name == "CreateTime" || field.Name == "UpdateTime" {
				// Validate they are time.Time type
				if field.Type != timeType {
					return fmt.Errorf("model field '%s' must be of type time.Time, got %s", field.Name, field.Type.String())
				}
			}

			modelFieldMap[fieldName] = field
		}
	}

	// Validate Count field type if it exists
	if modelHasCountField && countField.Type != countFieldType {
		return fmt.Errorf("model's Count field must be of type int64, got %s", countField.Type.String())
	}

	// Find missing fields
	var missingInTable []string
	var missingInModel []string

	// Check fields missing from table
	for modelFieldName := range modelFieldMap {
		if _, exists := tableFieldMap[modelFieldName]; !exists {
			missingInTable = append(missingInTable, modelFieldName)
		}
	}

	// Check fields missing from model
	for tableFieldName := range tableFieldMap {
		if _, exists := modelFieldMap[tableFieldName]; !exists {
			missingInModel = append(missingInModel, tableFieldName)
		}
	}

	// Report detailed field mismatch errors
	if len(missingInTable) > 0 || len(missingInModel) > 0 {
		var errMsg string
		if len(missingInTable) > 0 {
			errMsg += fmt.Sprintf("Fields in model but missing from table: %v", missingInTable)
			// Only add period if this is the only message
			if len(missingInModel) == 0 {
				errMsg += "."
			} else {
				errMsg += ". " // Add period and one space
			}
		}
		if len(missingInModel) > 0 {
			errMsg += fmt.Sprintf("Fields in table but missing from model: %v.", missingInModel)
		}
		return fmt.Errorf("%w: %s", ErrFieldCountMismatch, errMsg)
	}

	// Check field type compatibility for fields that exist in both
	for modelFieldName, structField := range modelFieldMap {
		tableField, exists := tableFieldMap[modelFieldName]
		if exists {
			if err := checkFieldTypeCompatibility(structField.Type, tableField); err != nil {
				return fmt.Errorf("%w: field %s, %v", ErrFieldTypeMismatch, modelFieldName, err)
			}
		}
	}

	return nil
}

// validateOptionalType checks if the optional fields type P is a struct
// with pointer fields that match the model type T.
func validateOptionalType[T, P any]() error {
	modelType := reflect.TypeOf((*T)(nil)).Elem()
	optionalType := reflect.TypeOf((*P)(nil)).Elem()
	timeType := reflect.TypeOf(time.Time{})

	// Check if P is a struct
	if optionalType.Kind() != reflect.Struct {
		return ErrNotPointerStruct
	}

	// Create a map of model fields for faster lookup
	modelFieldMap := make(map[string]reflect.StructField)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.IsExported() {
			modelFieldMap[field.Name] = field
		}
	}

	// Check each field in the optional type
	for i := 0; i < optionalType.NumField(); i++ {
		optField := optionalType.Field(i)

		// Skip unexported fields
		if !optField.IsExported() {
			continue
		}

		// Check time fields in optional struct
		if optField.Name == "CreateTime" || optField.Name == "UpdateTime" {
			// Must be a pointer
			if optField.Type.Kind() != reflect.Ptr {
				return fmt.Errorf("%w: optional field %s must be a pointer",
					ErrNotPointerStruct, optField.Name)
			}

			// Must be a *time.Time
			if optField.Type.Elem() != timeType {
				return fmt.Errorf("%w: optional field %s must be a *time.Time, got %s",
					ErrFieldTypeMismatch, optField.Name, optField.Type.String())
			}
		}

		// Check if field exists in model
		modelField, exists := modelFieldMap[optField.Name]
		if !exists {
			return fmt.Errorf("%w: optional field %s not found in model",
				ErrFieldMismatch, optField.Name)
		}

		// Check if field is a pointer
		if optField.Type.Kind() != reflect.Ptr {
			return fmt.Errorf("%w: optional field %s must be a pointer",
				ErrNotPointerStruct, optField.Name)
		}

		// Check if pointer type matches model field type
		if optField.Type.Elem() != modelField.Type {
			return fmt.Errorf("%w: optional field %s pointer type %s doesn't match model field type %s",
				ErrFieldTypeMismatch, optField.Name, optField.Type.Elem(), modelField.Type)
		}
	}

	return nil
}

// getFieldName extracts the field name from struct field or tags
// and converts it to the appropriate case for database fields
func getFieldName(field reflect.StructField) string {
	// Convert field name to snake_case for comparison with table fields
	return strcase.CamelToSnake(field.Name)
}

// hasConsecutiveUppercase checks if a string has two or more consecutive uppercase letters.
// This is used to enforce strict CamelCase naming (e.g., "SomeId" is valid, "SomeID" is not).
// Names like "ID" are also invalid - use "Id" instead.
func hasConsecutiveUppercase(s string) bool {
	prevUpper := false
	for _, r := range s {
		isUpper := r >= 'A' && r <= 'Z'
		if isUpper && prevUpper {
			return true
		}
		prevUpper = isUpper
	}
	return false
}

// toStrictCamelCase converts a field name to strict CamelCase format.
// It converts consecutive uppercase letters to have only the first letter uppercase,
// while preserving word boundaries.
// Examples:
//   - "SomeID" -> "SomeId"
//   - "SomeJSON" -> "SomeJson"
//   - "HTTPStatus" -> "HttpStatus"
//   - "ID" -> "Id"
//   - "URL" -> "Url"
//   - "HTTPSProtocol" -> "HttpsProtocol"
func toStrictCamelCase(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	result := make([]rune, len(runes))
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]
		isUpper := r >= 'A' && r <= 'Z'

		if isUpper && i > 0 {
			prevUpper := runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			if prevUpper {
				// Check if next character is lowercase (word boundary)
				// If next char is lowercase, keep this one uppercase (it's start of new word)
				// Otherwise, lowercase it
				nextIsLower := i+1 < n && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if !nextIsLower {
					// Convert consecutive uppercase to lowercase
					result[i] = r + ('a' - 'A')
					continue
				}
			}
		}
		result[i] = r
	}
	return string(result)
}

// validateFieldNaming checks if a field name follows strict CamelCase naming convention.
// Returns an error if the field name contains consecutive uppercase letters.
// All consecutive uppercase letters are invalid, including standalone acronyms like "ID".
// Use "Id" instead of "ID", "SomeId" instead of "SomeID", "SomeJson" instead of "SomeJSON".
func validateFieldNaming(fieldName string) error {
	if hasConsecutiveUppercase(fieldName) {
		corrected := toStrictCamelCase(fieldName)
		return fmt.Errorf("%w: field '%s' has consecutive uppercase letters, use '%s' instead",
			ErrInvalidFieldNaming, fieldName, corrected)
	}
	return nil
}

// checkFieldTypeCompatibility checks if a struct field type is compatible with a table field
func checkFieldTypeCompatibility(structType reflect.Type, tableField field.Field) error {
	// 1. Check if the Go type is compatible with the database type
	// 2. Handle conversions between related types (e.g. int64 and int)
	// NOTE: DB int can be converted to bool
	switch tableField.(type) {
	case field.Int64Field:
		if structType.Kind() != reflect.Int64 && structType.Kind() != reflect.Int && structType.Kind() != reflect.Bool {
			return fmt.Errorf("expected int/int64 for Int64Field, got %s", structType.String())
		}
	case field.Int32Field:
		if structType.Kind() != reflect.Int32 && structType.Kind() != reflect.Int && structType.Kind() != reflect.Bool {
			return fmt.Errorf("expected int32 for Int32Field, got %s", structType.String())
		}
	case field.StringField:
		if structType.Kind() != reflect.String {
			return fmt.Errorf("expected string for StringField, got %s", structType.String())
		}
	case field.TimeField:
		// Time is a struct, so check against time.Time type name
		if structType.String() != "time.Time" {
			return fmt.Errorf("expected time.Time for TimeField, got %s", structType.String())
		}
	case field.Float64Field:
		if structType.Kind() != reflect.Float64 {
			return fmt.Errorf("expected float64 for Float64Field, got %s", structType.String())
		}
	default:
		return fmt.Errorf("unsupported table field type: %T", tableField)
	}

	return nil
}
