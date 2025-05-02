package orm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/ormx/engine"
	"github.com/xhd2015/ormx/table"
	"github.com/xhd2015/xgo/support/assert"
)

// MockQueryEngine implements engine.Engine for testing queries
type MockQueryEngine struct {
	MockEngine
	// QueryFunc allows customizing the behavior of QueryMany
	QueryFunc func(ctx context.Context, sql string, args []interface{}, result interface{}) error
}

func (m *MockQueryEngine) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args, result)
	}
	return nil
}

func (m *MockQueryEngine) GetEngine() engine.Engine {
	return m
}

// MockEngine is a minimal engine implementation
type MockEngine struct {
	// Add fields for tracking method calls
	ExecCalls       []ExecCall
	ExecInsertCalls []ExecInsertCall
}

type ExecCall struct {
	SQL  string
	Args []interface{}
}

type ExecInsertCall struct {
	SQL  string
	Args []interface{}
}

func (m *MockEngine) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	// Default implementation that does nothing
	return nil
}

func (m *MockEngine) Exec(ctx context.Context, sql string, args []interface{}) error {
	// Track the call
	m.ExecCalls = append(m.ExecCalls, ExecCall{SQL: sql, Args: args})
	return nil
}

func (m *MockEngine) ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error) {
	// Track the call
	m.ExecInsertCalls = append(m.ExecInsertCalls, ExecInsertCall{SQL: sql, Args: args})
	return 42, nil // Return a dummy ID
}

func (m *MockEngine) GetEngine() engine.Engine {
	return m
}

// TestModel for query tests
type TestModel struct {
	ID    int64
	Name  string
	Age   int
	Count int64 // Added for Count tests
}

// TestModelOptional for optional fields in tests
type TestModelOptional struct {
	ID    *int64
	Name  *string
	Age   *int
	Count *int64 // Added for Count tests
}

// TestModel for query tests - extended with time fields
type TestModelWithTime struct {
	ID         int64
	Name       string
	Age        int
	Count      int64
	CreateTime time.Time
	UpdateTime time.Time
}

// Optional type for TestModelWithTime
type TestModelWithTimeOptional struct {
	ID         *int64
	Name       *string
	Age        *int
	Count      *int64
	CreateTime *time.Time
	UpdateTime *time.Time
}

func TestQuery_Success(t *testing.T) {
	// Setup a mock engine that returns test data
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			// Check that the query and args were correctly passed
			if sql != "SELECT * FROM test_table WHERE age > ?" {
				t.Errorf("Expected query 'SELECT * FROM test_table WHERE age > ?', got %s", sql)
			}

			if len(args) != 1 || args[0] != 18 {
				t.Errorf("Expected args [18], got %v", args)
			}

			// Check that result is a pointer to a slice of *TestModel
			resultPtr, ok := result.(*[]*TestModel)
			if !ok {
				t.Fatalf("Expected result to be *[]*TestModel, got %T", result)
			}

			// Populate the result with test data
			*resultPtr = []*TestModel{
				{ID: 1, Name: "Alice", Age: 25, Count: 0},
				{ID: 2, Name: "Bob", Age: 30, Count: 0},
			}

			return nil
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	// No count field in table

	// Create ORM instance
	orm, err := bind[TestModel, TestModelOptional](mockEngine, testTable)
	if err != nil {
		t.Fatalf("Failed to create ORM: %v", err)
	}

	// Execute a query
	query := "SELECT * FROM test_table WHERE age > ?"
	args := []interface{}{18}

	results, err := orm.Query(context.Background(), query, args)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Check first result
	if results[0].ID != 1 || results[0].Name != "Alice" || results[0].Age != 25 {
		t.Errorf("First result does not match expected data: %+v", results[0])
	}

	// Check second result
	if results[1].ID != 2 || results[1].Name != "Bob" || results[1].Age != 30 {
		t.Errorf("Second result does not match expected data: %+v", results[1])
	}
}

func TestQuery_Error(t *testing.T) {
	// Setup a mock engine that returns an error
	expectedErr := errors.New("database error")
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			return expectedErr
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	// No count field in table

	// Create ORM instance
	orm, err := bind[TestModel, TestModelOptional](mockEngine, testTable)
	if err != nil {
		t.Fatalf("Failed to create ORM: %v", err)
	}

	// Execute a query
	results, err := orm.Query(context.Background(), "SELECT * FROM test_table", nil)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}

	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

func TestQuery_EmptyResults(t *testing.T) {
	// Setup a mock engine that returns no results
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			// Return empty result - don't modify the slice
			return nil
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	// No count field in table

	// Create ORM instance
	orm, err := bind[TestModel, TestModelOptional](mockEngine, testTable)
	if err != nil {
		t.Fatalf("Failed to create ORM: %v", err)
	}

	// Execute a query
	results, err := orm.Query(context.Background(), "SELECT * FROM test_table", nil)

	// Verify empty results handling
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestQueryByID_Found(t *testing.T) {
	// Setup a mock engine that returns a result
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			// Check that the query and args were correctly passed
			expectedSQL := "SELECT `test_table`.`id`, `test_table`.`name`, `test_table`.`age` FROM `test_table` WHERE `test_table`.`id` = ? LIMIT 1"
			if sql != expectedSQL {
				t.Errorf("Expected query '%s', got '%s'", expectedSQL, sql)
			}

			if len(args) != 1 || args[0] != int64(42) {
				t.Errorf("Expected args [42], got %v", args)
			}

			// Check that result is a pointer to a slice of *TestModel
			resultPtr, ok := result.(*[]*TestModel)
			if !ok {
				t.Fatalf("Expected result to be *[]*TestModel, got %T", result)
			}

			// Populate the result with test data
			*resultPtr = []*TestModel{
				{ID: 42, Name: "Alice", Age: 25},
			}

			return nil
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute QueryByID
	result, err := orm.GetByID(context.Background(), 42)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected a result, got nil")
	}

	if result.ID != 42 || result.Name != "Alice" || result.Age != 25 {
		t.Errorf("Result does not match expected data: %+v", result)
	}
}

func TestQueryByID_NotFound(t *testing.T) {
	// Setup a mock engine that returns no results
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			// Return empty result
			resultPtr, ok := result.(*[]*TestModel)
			if !ok {
				t.Fatalf("Expected result to be *[]*TestModel, got %T", result)
			}

			*resultPtr = []*TestModel{} // Empty slice
			return nil
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute QueryByID
	_, err := orm.GetByID(context.Background(), 99)

	// Verify not found handling
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if diff := assert.Diff(err.Error(), "test_table not found with: id=99"); diff != "" {
		t.Error(diff)
	}
}

func TestQueryByID_Error(t *testing.T) {
	// Setup a mock engine that returns an error
	expectedErr := errors.New("database error")
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			return expectedErr
		},
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute QueryByID
	result, err := orm.GetByID(context.Background(), 42)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Errorf("Expected nil result for error case, got %+v", result)
	}
}

func TestInsert(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Model to insert
	model := &TestModel{
		ID:   123,
		Name: "Charlie",
		Age:  35,
	}

	// Execute Insert
	id, err := orm.Insert(context.Background(), model)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecInsertCalls) != 1 {
		t.Fatalf("Expected 1 ExecInsert call, got %d", len(mockEngine.ExecInsertCalls))
	}

	call := mockEngine.ExecInsertCalls[0]

	// Verify the generated SQL follows the INSERT INTO SET format
	if !strings.HasPrefix(call.SQL, "INSERT INTO `test_table` SET ") {
		t.Errorf("Expected SQL to start with 'INSERT INTO `test_table` SET ', got %q", call.SQL)
	}

	// Check for field-value pairs in the SQL
	expectedFieldSets := []string{
		"`id`=?",
		"`name`=?",
		"`age`=?",
	}

	for _, fieldSet := range expectedFieldSets {
		if !strings.Contains(call.SQL, fieldSet) {
			t.Errorf("Expected SQL to contain %q, got %q", fieldSet, call.SQL)
		}
	}

	// Verify the args contain the values from our model
	if len(call.Args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(call.Args))
	} else {
		// The args order should match the order of the SET clauses
		// But since map iteration order is not guaranteed, we'll check that they're all there
		containsID := false
		containsName := false
		containsAge := false

		for _, arg := range call.Args {
			switch v := arg.(type) {
			case int64:
				if v == 123 {
					containsID = true
				} else if v == 35 {
					containsAge = true
				}
			case string:
				if v == "Charlie" {
					containsName = true
				}
			}
		}

		if !containsID {
			t.Errorf("Args do not contain ID value")
		}
		if !containsName {
			t.Errorf("Args do not contain Name value")
		}
		if !containsAge {
			t.Errorf("Args do not contain Age value")
		}
	}
}

func TestUpdateByID(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Fields to update
	name := "Updated Name"
	partialModel := &TestModelOptional{
		Name: &name,
	}

	// Execute UpdateByID
	err := orm.UpdateByID(context.Background(), 42, partialModel)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecCalls) != 1 {
		t.Fatalf("Expected 1 Exec call, got %d", len(mockEngine.ExecCalls))
	}

	call := mockEngine.ExecCalls[0]
	expectedSQL := "UPDATE `test_table` SET `name`=? WHERE `id` = ?"
	if diff := assert.Diff(call.SQL, expectedSQL); diff != "" {
		t.Error(diff)
	}

	if len(call.Args) != 2 || call.Args[0] != "Updated Name" || call.Args[1] != int64(42) {
		t.Errorf("Expected args [%q, 42], got %v", "Updated Name", call.Args)
	}
}

// Test for updating multiple fields
func TestUpdateByID_MultipleFields(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Fields to update
	name := "Updated Name"
	age := 30
	partialModel := &TestModelOptional{
		Name: &name,
		Age:  &age,
	}

	// Execute UpdateByID
	err := orm.UpdateByID(context.Background(), 42, partialModel)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecCalls) != 1 {
		t.Fatalf("Expected 1 Exec call, got %d", len(mockEngine.ExecCalls))
	}

	call := mockEngine.ExecCalls[0]
	// Since map iteration order is not guaranteed, we can't test the exact SQL string
	// Instead, check that it contains both fields
	expectedSQL := "UPDATE `test_table` SET `name`=?, `age`=? WHERE `id` = ?"
	if diff := assert.Diff(call.SQL, expectedSQL); diff != "" {
		t.Error(diff)
	}

	args := fmt.Sprint(call.Args)

	if diff := assert.Diff(args, "[Updated Name 30 42]"); diff != "" {
		t.Error(diff)
	}
}

// Test for nothing to update
func TestUpdateByID_NothingToUpdate(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Empty partial model
	partialModel := &TestModelOptional{}

	// Execute UpdateByID
	err := orm.UpdateByID(context.Background(), 42, partialModel)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, ErrNothingToUpdate) {
		t.Errorf("Expected error to be ErrNothingToUpdate, got %v", err)
	}

	// Verify no calls were made to the engine
	if len(mockEngine.ExecCalls) != 0 {
		t.Errorf("Expected 0 Exec calls, got %d", len(mockEngine.ExecCalls))
	}
}

// Test for missing ID field
func TestUpdateByID_MissingIDField(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table WITHOUT the ID field
	testTable := table.New("test_table")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Fields to update
	name := "Updated Name"
	partialModel := &TestModelOptional{
		Name: &name,
	}

	// Execute UpdateByID
	err := orm.UpdateByID(context.Background(), 42, partialModel)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, ErrMissingIDField) {
		t.Errorf("Expected error to be ErrMissingIDField, got %v", err)
	}

	// Verify no calls were made to the engine
	if len(mockEngine.ExecCalls) != 0 {
		t.Errorf("Expected 0 Exec calls, got %d", len(mockEngine.ExecCalls))
	}
}

// Test for Count with count field
func TestCount_Success(t *testing.T) {
	// Setup a mock engine that returns test data
	mockEngine := &MockQueryEngine{
		QueryFunc: func(ctx context.Context, sql string, args []interface{}, result interface{}) error {
			// Check that the query and args were correctly passed
			if sql != "SELECT COUNT(*) as count FROM test_table" {
				t.Errorf("Expected query 'SELECT COUNT(*) as count FROM test_table', got %s", sql)
			}

			// Check that result is a pointer to a slice of *TestModel
			resultPtr, ok := result.(*[]*TestModel)
			if !ok {
				t.Fatalf("Expected result to be *[]*TestModel, got %T", result)
			}

			// Populate the result with test data - a single row with count value
			*resultPtr = []*TestModel{
				{ID: 0, Name: "", Age: 0, Count: 5}, // Use Count field for the count value
			}

			return nil
		},
	}

	// Create a test table without count field
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	// No count field in table

	// Create ORM instance
	orm, err := bind[TestModel, TestModelOptional](mockEngine, testTable)
	if err != nil {
		t.Fatalf("Failed to create ORM: %v", err)
	}

	// Execute a count query
	query := "SELECT COUNT(*) as count FROM test_table"
	results, err := orm.Count(context.Background(), query, nil)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Check the Count field value
	if results[0].Count != 5 {
		t.Errorf("Expected count 5, got %d", results[0].Count)
	}
}

// Test to verify validation rejects tables with 'count' field
func TestValidate_RejectsCountField(t *testing.T) {
	// Create a test table WITH a count field
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	testTable.Int64("count") // This should cause validation to fail

	// Try to create ORM instance
	_, err := bind[TestModel, TestModelOptional](nil, testTable)

	// Verify error
	if err == nil {
		t.Fatal("Expected an error for table with count field, got nil")
	}

	if !strings.Contains(err.Error(), "table must not contain a 'count' field") {
		t.Errorf("Expected error about count field, got: %v", err)
	}
}

// Test to verify model with wrong Count field type is rejected
func TestValidate_WrongCountFieldType(t *testing.T) {
	// Define a model with wrong Count field type
	type WrongCountModel struct {
		ID    int64
		Name  string
		Age   int
		Count string // Should be int64
	}

	type WrongCountOptional struct {
		ID    *int64
		Name  *string
		Age   *int
		Count *string
	}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Try to create ORM instance
	_, err := bind[WrongCountModel, WrongCountOptional](nil, testTable)

	// Verify error
	if err == nil {
		t.Fatal("Expected an error for model with wrong Count field type, got nil")
	}

	if !strings.Contains(err.Error(), "model's Count field must be of type int64") {
		t.Errorf("Expected error about Count field type, got: %v", err)
	}
}

// Test for Count with model lacking Count field
func TestCount_ModelLacksCountField(t *testing.T) {
	// Define a model without Count field
	type NoCountModel struct {
		ID   int64
		Name string
		Age  int
		// No Count field
	}

	type NoCountOptional struct {
		ID   *int64
		Name *string
		Age  *int
	}

	// Setup a mock engine
	mockEngine := &MockQueryEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly to bypass validation
	orm := &ORM[NoCountModel, NoCountOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute Count
	results, err := orm.Count(context.Background(), "SELECT COUNT(*) FROM test_table", nil)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error for model without Count field, got nil")
	}

	if !errors.Is(err, ErrMissingCountField) {
		t.Errorf("Expected error to be ErrMissingCountField, got %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

// Test for Count with model having wrong Count field type
func TestCount_WrongCountFieldType(t *testing.T) {
	// Define a model with wrong Count field type
	type WrongCountTypeModel struct {
		ID    int64
		Name  string
		Age   int
		Count string // Wrong type, should be int64
	}

	type WrongCountTypeOptional struct {
		ID    *int64
		Name  *string
		Age   *int
		Count *string
	}

	// Setup a mock engine
	mockEngine := &MockQueryEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly to bypass validation
	orm := &ORM[WrongCountTypeModel, WrongCountTypeOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute Count
	results, err := orm.Count(context.Background(), "SELECT COUNT(*) FROM test_table", nil)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error for model with wrong Count field type, got nil")
	}

	if !strings.Contains(err.Error(), "model type must have a Count field of type int64") {
		t.Errorf("Expected error about Count field type, got: %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

// TestInsertWithTimeFields tests the automatic setting of time fields
func TestInsertWithTimeFields(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	testTable.Time("create_time")
	testTable.Time("update_time")

	// Create ORM instance directly
	orm := &ORM[TestModelWithTime, TestModelWithTimeOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Model to insert with zero time fields
	model := &TestModelWithTime{
		ID:   123,
		Name: "Charlie",
		Age:  35,
		// CreateTime and UpdateTime are zero values
	}

	// Record the current time to compare
	beforeInsert := time.Now()

	// Execute Insert
	id, err := orm.Insert(context.Background(), model)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecInsertCalls) != 1 {
		t.Fatalf("Expected 1 ExecInsert call, got %d", len(mockEngine.ExecInsertCalls))
	}

	call := mockEngine.ExecInsertCalls[0]

	// Check that all necessary fields are in the SQL
	if !strings.Contains(call.SQL, "`create_time`=?") || !strings.Contains(call.SQL, "`update_time`=?") {
		t.Errorf("Expected SQL to include time fields, got %q", call.SQL)
	}

	// Verify we have at least 5 arguments (id, name, age, create_time, update_time)
	if len(call.Args) < 5 {
		t.Errorf("Expected at least 5 args, got %d", len(call.Args))
		return
	}

	// Try to find the time values in the args
	var foundCreateTime, foundUpdateTime bool
	var createTime, updateTime time.Time

	for _, arg := range call.Args {
		if timeArg, ok := arg.(time.Time); ok {
			// Should be after our beforeInsert time and before now
			if timeArg.After(beforeInsert) && timeArg.Before(time.Now().Add(time.Second)) {
				// This is a valid auto-generated time
				// Try to determine if it's create or update time based on order
				if !foundCreateTime {
					createTime = timeArg
					foundCreateTime = true
				} else if !foundUpdateTime {
					updateTime = timeArg
					foundUpdateTime = true
				}
			}
		}
	}

	if !foundCreateTime {
		t.Errorf("CreateTime was not automatically set")
	}

	if !foundUpdateTime {
		t.Errorf("UpdateTime was not automatically set")
	}

	// They should be very close together
	if foundCreateTime && foundUpdateTime {
		timeDiff := updateTime.Sub(createTime).Abs()
		if timeDiff > time.Millisecond*100 {
			t.Errorf("CreateTime and UpdateTime differ by too much: %v", timeDiff)
		}
	}
}

// TestUpdateByID_AutoUpdateTime tests the automatic setting of UpdateTime field
func TestUpdateByID_AutoUpdateTime(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")
	testTable.Time("update_time")

	// Create ORM instance directly
	orm := &ORM[TestModelWithTime, TestModelWithTimeOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Fields to update - note that UpdateTime is nil
	name := "Updated Name"
	partialModel := &TestModelWithTimeOptional{
		Name: &name,
		// UpdateTime is intentionally nil
	}

	// Record the current time to compare
	beforeUpdate := time.Now()

	// Execute UpdateByID
	err := orm.UpdateByID(context.Background(), 42, partialModel)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecCalls) != 1 {
		t.Fatalf("Expected 1 Exec call, got %d", len(mockEngine.ExecCalls))
	}

	call := mockEngine.ExecCalls[0]

	// Check that the SQL includes both fields
	if diff := assert.Diff(call.SQL, "UPDATE `test_table` SET `name`=?, `update_time`=? WHERE `id` = ?"); diff != "" {
		t.Error(diff)
	}

	// Verify we have at least 3 arguments (name, update_time, id)
	if len(call.Args) < 3 {
		t.Errorf("Expected at least 3 args, got %d", len(call.Args))
		return
	}

	// Try to find the time value in the args
	var foundUpdateTime bool

	for _, arg := range call.Args {
		if timeArg, ok := arg.(time.Time); ok {
			// Should be after our beforeUpdate time and before now
			if timeArg.After(beforeUpdate) && timeArg.Before(time.Now().Add(time.Second)) {
				foundUpdateTime = true
				break
			}
		}
	}

	if !foundUpdateTime {
		t.Errorf("UpdateTime was not automatically set")
	}
}

// TestDeleteByID tests deleting a record by ID
func TestDeleteByID(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table
	testTable := table.New("test_table")
	testTable.Int64("id")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute DeleteByID
	err := orm.DeleteByID(context.Background(), 42)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the correct call was made to the engine
	if len(mockEngine.ExecCalls) != 1 {
		t.Fatalf("Expected 1 Exec call, got %d", len(mockEngine.ExecCalls))
	}

	call := mockEngine.ExecCalls[0]
	expectedSQL := "DELETE FROM `test_table` WHERE `id` = ?"
	if call.SQL != expectedSQL {
		t.Errorf("Expected SQL %q, got %q", expectedSQL, call.SQL)
	}

	if len(call.Args) != 1 || call.Args[0] != int64(42) {
		t.Errorf("Expected args [42], got %v", call.Args)
	}
}

// TestDeleteByID_MissingIDField tests handling when the table is missing an ID field
func TestDeleteByID_MissingIDField(t *testing.T) {
	// Setup a mock engine
	mockEngine := &MockEngine{}

	// Create a test table WITHOUT the ID field
	testTable := table.New("test_table")
	testTable.String("name")
	testTable.Int64("age")

	// Create ORM instance directly
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute DeleteByID
	err := orm.DeleteByID(context.Background(), 42)

	// Verify error handling
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, ErrMissingIDField) {
		t.Errorf("Expected error to be ErrMissingIDField, got %v", err)
	}

	// Verify no calls were made to the engine
	if len(mockEngine.ExecCalls) != 0 {
		t.Errorf("Expected 0 Exec calls, got %d", len(mockEngine.ExecCalls))
	}
}
