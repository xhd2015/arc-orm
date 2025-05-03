package orm

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/arc-orm/engine"
	"github.com/xhd2015/arc-orm/table"
	"github.com/xhd2015/less-gen/strcase"
	"github.com/xhd2015/xgo/support/assert"
)

// Mock engine for testing
type mockEngine struct{}

func (m *mockEngine) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	return nil
}

func (m *mockEngine) Exec(ctx context.Context, sql string, args []interface{}) error {
	return nil
}

func (m *mockEngine) ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error) {
	return 0, nil
}

func (m *mockEngine) GetEngine() engine.Engine {
	return m
}

// Valid test model
type ValidModel struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}

// Valid optional fields for the valid model
type ValidOptional struct {
	ID        *int64
	Name      *string
	Email     *string
	CreatedAt *time.Time
}

// Model with extra field
type ModelWithExtraField struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	Extra     bool // Extra field not in table
}

// Model missing a field
type ModelMissingField struct {
	ID   int64
	Name string
	// Email field missing
	CreatedAt time.Time
}

// Model with wrong field type
type ModelWrongFieldType struct {
	ID        int64
	Name      string
	Email     int64 // Should be string
	CreatedAt time.Time
}

// Optional fields with non-pointer field
type OptionalNonPointer struct {
	ID        *int64
	Name      string // Should be *string
	Email     *string
	CreatedAt *time.Time
}

// Create a valid table for testing
func createValidTable() table.Table {
	t := table.New("users")
	// Table field names should match the snake_case versions of model field names
	t.Int64("id")
	t.String("name")
	t.String("email")
	t.Time("created_at")
	return t
}

// Create a table missing a field
func createTableMissingField() table.Table {
	t := table.New("users")
	t.Int64("id")
	t.String("name")
	// Email field missing
	t.Time("created_at")
	return t
}

func TestValidate_ValidModel(t *testing.T) {
	// Setup
	validTable := createValidTable()
	mockEng := &mockEngine{}

	// Test with valid model and options
	orm, err := bind[ValidModel, ValidOptional](mockEng, validTable)
	if err != nil {
		t.Fatalf("Expected validation to pass but got error: %v", err)
	}
	if orm == nil {
		t.Fatal("Expected ORM instance to be returned")
	}
}

func TestValidate_ModelWithExtraField(t *testing.T) {
	// Setup
	validTable := createValidTable()
	mockEng := &mockEngine{}

	// Test with model having extra field
	_, err := bind[ModelWithExtraField, ValidOptional](mockEng, validTable)
	if err == nil {
		t.Fatal("Expected validation to fail due to extra field, but it passed")
	}

	expectedErrMsg := "ORM validation failed: model validation failed: number of fields in model does not match table: Fields in model but missing from table: [extra]."
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

func TestValidate_ModelMissingField(t *testing.T) {
	// Setup
	validTable := createValidTable()
	mockEng := &mockEngine{}

	// Test with model missing a field
	_, err := bind[ModelMissingField, ValidOptional](mockEng, validTable)
	if err == nil {
		t.Fatal("Expected validation to fail due to missing field, but it passed")
	}

	expectedErrMsg := "ORM validation failed: model validation failed: number of fields in model does not match table: Fields in table but missing from model: [email]."
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

func TestValidate_TableMissingField(t *testing.T) {
	// Setup
	tableMissingField := createTableMissingField()
	mockEng := &mockEngine{}

	// Test with table missing a field
	_, err := bind[ValidModel, ValidOptional](mockEng, tableMissingField)
	if err == nil {
		t.Fatal("Expected validation to fail due to table missing field, but it passed")
	}

	expectedErrMsg := "ORM validation failed: model validation failed: number of fields in model does not match table: Fields in model but missing from table: [email]."
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Error(diff)
	}
}

func TestValidate_WrongFieldType(t *testing.T) {
	// Setup
	validTable := createValidTable()
	mockEng := &mockEngine{}

	// Test with model having wrong field type
	_, err := bind[ModelWrongFieldType, ValidOptional](mockEng, validTable)
	if err == nil {
		t.Fatal("Expected validation to fail due to wrong field type, but it passed")
	}

	expectedErrMsg := "ORM validation failed: model validation failed: field type mismatch between model and table: field email, expected string for StringField, got int64"
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

func TestValidate_OptionalNonPointer(t *testing.T) {
	// Setup
	validTable := createValidTable()
	mockEng := &mockEngine{}

	// Test with optional fields having non-pointer field
	_, err := bind[ValidModel, OptionalNonPointer](mockEng, validTable)
	if err == nil {
		t.Fatal("Expected validation to fail due to non-pointer optional field, but it passed")
	}

	expectedErrMsg := "ORM validation failed: optional fields validation failed: optional fields type must be a struct with pointer fields: optional field Name must be a pointer"
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

// Add these new tests at the bottom of the file before the helper functions

func TestValidate_NonStructModel(t *testing.T) {
	// Using string as a non-struct model type
	_, err := bind[string, ValidOptional](&mockEngine{}, createValidTable())
	if err == nil {
		t.Fatal("Expected validation to fail for non-struct model, but it passed")
	}

	expectedErrMsg := "ORM validation failed: model validation failed: model type must be a struct"
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

func TestValidate_NonStructOptional(t *testing.T) {
	// Using string as a non-struct optional type
	_, err := bind[ValidModel, string](&mockEngine{}, createValidTable())
	if err == nil {
		t.Fatal("Expected validation to fail for non-struct optional fields, but it passed")
	}

	expectedErrMsg := "ORM validation failed: optional fields validation failed: optional fields type must be a struct with pointer fields"
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

func TestValidate_DirectCall(t *testing.T) {
	// Create ORM without validation
	orm := &ORM[ValidModel, OptionalNonPointer]{
		table:  createValidTable(),
		engine: &mockEngine{},
	}

	// Call Validate directly
	err := orm.Validate()
	if err == nil {
		t.Fatal("Expected direct validation call to fail due to non-pointer optional field, but it passed")
	}

	expectedErrMsg := "optional fields validation failed: optional fields type must be a struct with pointer fields: optional field Name must be a pointer"
	if diff := assert.Diff(expectedErrMsg, err.Error()); diff != "" {
		t.Errorf("Expected error message:\n%s\nGot:\n%s", expectedErrMsg, err.Error())
	}
}

// Test for field name conversion
func TestValidate_FieldNameConversion(t *testing.T) {
	// Create a struct with a field name that needs conversion
	type ModelWithCamelCase struct {
		ID           int64
		UserName     string // This should be converted to "user_name" in snake_case
		EmailAddress string // This should be converted to "email_address" in snake_case
		CreatedAt    time.Time
	}

	// Create optional fields with matching names
	type CamelCaseOptional struct {
		ID           *int64
		UserName     *string
		EmailAddress *string
		CreatedAt    *time.Time
	}

	// Create a table with snake_case field names matching the converted struct field names
	table := table.New("users")
	table.Int64("id")
	table.String("user_name")     // Must match snake_case of UserName
	table.String("email_address") // Must match snake_case of EmailAddress
	table.Time("created_at")      // Must match snake_case of CreatedAt

	// Verify our conversion logic works as expected
	userNameField := reflect.StructField{Name: "UserName"}
	convertedName := getFieldName(userNameField)
	if convertedName != "user_name" {
		t.Errorf("Field name conversion failed. Expected 'user_name', got '%s'", convertedName)
	}

	// Check if ORM validation succeeds with proper conversion
	orm, err := bind[ModelWithCamelCase, CamelCaseOptional](&mockEngine{}, table)
	if err != nil {
		t.Fatalf("Expected validation to pass with name conversion but got error: %v", err)
	}
	if orm == nil {
		t.Fatal("Expected ORM instance to be returned")
	}

	// Check direct conversion matches what we expect
	if strcase.CamelToSnake("UserName") != "user_name" {
		t.Errorf("CamelToSnake conversion unexpected. Expected 'user_name', got '%s'", strcase.CamelToSnake("UserName"))
	}
}

// TestValidate_TimeFields tests validation of create_time and update_time fields
func TestValidate_TimeFields(t *testing.T) {
	// Create a test table with proper time fields
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Time("create_time")
	testTable.Time("update_time")

	// Define a model with proper time fields
	type ModelWithTimeFields struct {
		ID         int64
		Name       string
		CreateTime time.Time
		UpdateTime time.Time
	}

	type ModelWithTimeFieldsOpt struct {
		ID         *int64
		Name       *string
		CreateTime *time.Time
		UpdateTime *time.Time
	}

	// Test with proper types - should pass
	_, err := bind[ModelWithTimeFields, ModelWithTimeFieldsOpt](nil, testTable)
	if err != nil {
		t.Errorf("Expected validation to pass for correct time fields, got error: %v", err)
	}
}

// TestValidate_WrongTimeTypes tests validation with incorrect time field types
func TestValidate_WrongTimeTypes(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		testSetup     func() (table.Table, error)
		expectedError string
	}{
		{
			name: "Model with wrong CreateTime type",
			testSetup: func() (table.Table, error) {
				testTable := table.New("test_table")
				testTable.Int64("id")
				testTable.Time("create_time")
				testTable.Time("update_time")

				type WrongCreateTimeType struct {
					ID         int64
					CreateTime string // should be time.Time
					UpdateTime time.Time
				}

				type WrongCreateTimeTypeOpt struct {
					ID         *int64
					CreateTime *string
					UpdateTime *time.Time
				}

				_, err := bind[WrongCreateTimeType, WrongCreateTimeTypeOpt](nil, testTable)
				return testTable, err
			},
			expectedError: "model field 'CreateTime' must be of type time.Time",
		},
		{
			name: "Model with wrong UpdateTime type",
			testSetup: func() (table.Table, error) {
				testTable := table.New("test_table")
				testTable.Int64("id")
				testTable.Time("create_time")
				testTable.Time("update_time")

				type WrongUpdateTimeType struct {
					ID         int64
					CreateTime time.Time
					UpdateTime int64 // should be time.Time
				}

				type WrongUpdateTimeTypeOpt struct {
					ID         *int64
					CreateTime *time.Time
					UpdateTime *int64
				}

				_, err := bind[WrongUpdateTimeType, WrongUpdateTimeTypeOpt](nil, testTable)
				return testTable, err
			},
			expectedError: "model field 'UpdateTime' must be of type time.Time",
		},
		{
			name: "Optional with wrong CreateTime type",
			testSetup: func() (table.Table, error) {
				testTable := table.New("test_table")
				testTable.Int64("id")
				testTable.Time("create_time")
				testTable.Time("update_time")

				type Model struct {
					ID         int64
					CreateTime time.Time
					UpdateTime time.Time
				}

				type WrongOptionalType struct {
					ID         *int64
					CreateTime *string // should be *time.Time
					UpdateTime *time.Time
				}

				_, err := bind[Model, WrongOptionalType](nil, testTable)
				return testTable, err
			},
			expectedError: "optional field CreateTime must be a *time.Time",
		},
		{
			name: "Table with wrong create_time type",
			testSetup: func() (table.Table, error) {
				testTable := table.New("test_table")
				testTable.Int64("id")
				testTable.Int64("create_time") // should be Time
				testTable.Time("update_time")

				type Model struct {
					ID         int64
					CreateTime time.Time
					UpdateTime time.Time
				}

				type OptModel struct {
					ID         *int64
					CreateTime *time.Time
					UpdateTime *time.Time
				}

				_, err := bind[Model, OptModel](nil, testTable)
				return testTable, err
			},
			expectedError: "table field 'create_time' must be of type TimeField",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.testSetup()
			if err == nil {
				t.Fatalf("Expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tc.expectedError, err)
			}
		})
	}
}
