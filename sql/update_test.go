package sql

import (
	"testing"
)

func TestUpdateQueries(t *testing.T) {
	// Test basic UPDATE query
	query := Update(userTable.Name()).
		Set(UserName, String("John Doe")).
		Where(UserID.Eq(1))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `name`=? WHERE `users`.`id` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}
	if params[0] != "John Doe" {
		t.Errorf("Expected first param to be 'John Doe', got %v", params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 1 {
		t.Errorf("Expected second param to be int64(1), got %T %v", params[1], params[1])
	}

	// Test increment with field method
	emailUpdateQuery := Update(userTable.Name()).
		Set(UserEmail, UserEmail.Concat("addition")).
		Where(UserID.Eq(1))

	sqlStr, params, err = emailUpdateQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedIncrementSQL := "UPDATE `users` SET `email`=`users`.`email`+? WHERE `users`.`id` = ?"
	if sqlStr != expectedIncrementSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedIncrementSQL, sqlStr)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}
	if params[0] != "addition" {
		t.Errorf("Expected first param to be 'addition', got %v", params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 1 {
		t.Errorf("Expected second param to be int64(1), got %T %v", params[1], params[1])
	}

	// Test multiple SET expressions
	multiUpdateQuery := Update(userTable.Name()).
		Set(UserName, String("Jane Doe")).
		Set(UserEmail, UserEmail.Concat("subtraction")).
		Where(UserID.Eq(1))

	sqlStr, params, err = multiUpdateQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedMultiSQL := "UPDATE `users` SET `name`=?, `email`=`users`.`email`+? WHERE `users`.`id` = ?"
	if sqlStr != expectedMultiSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedMultiSQL, sqlStr)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}
	if params[0] != "Jane Doe" {
		t.Errorf("Expected first param to be 'Jane Doe', got %v", params[0])
	}
	if params[1] != "subtraction" {
		t.Errorf("Expected second param to be 'subtraction', got %v", params[1])
	}
	if v, ok := params[2].(int64); !ok || v != 1 {
		t.Errorf("Expected third param to be int64(1), got %T %v", params[2], params[2])
	}

	// Test the specific case for UPDATE `users` SET `age`=`age`+1 WHERE `id` = 1
	incrementAgeQuery := Update(userTable.Name()).
		Set(UserAge, UserAge.Increment(1)).
		Where(UserID.Eq(1))

	sqlStr, params, err = incrementAgeQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedAgeSQL := "UPDATE `users` SET `age`=`users`.`age`+? WHERE `users`.`id` = ?"
	if sqlStr != expectedAgeSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedAgeSQL, sqlStr)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 1 {
		t.Errorf("Expected first param to be int64(1), got %T %v", params[0], params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 1 {
		t.Errorf("Expected second param to be int64(1), got %T %v", params[1], params[1])
	}
}

func TestReadmeUpdateExample(t *testing.T) {
	// Test the UPDATE example from README.md
	query := Update(userTable.Name()).
		Set(UserName, String("John Doe")).
		Set(UserAge, UserAge.Increment(1)).
		Where(UserID.Eq(1))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `name`=?, `age`=`users`.`age`+? WHERE `users`.`id` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}
	if params[0] != "John Doe" {
		t.Errorf("Expected first param to be 'John Doe', got %v", params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 1 {
		t.Errorf("Expected second param to be int64(1), got %T %v", params[1], params[1])
	}
	if v, ok := params[2].(int64); !ok || v != 1 {
		t.Errorf("Expected third param to be int64(1), got %T %v", params[2], params[2])
	}
}

func TestUpdateBuilder_ValidationErrors(t *testing.T) {
	// Test missing table name
	query := Update("").
		Set(UserName, String("John"))
	_, _, err := query.SQL()
	if err == nil {
		t.Error("expected error for missing table name")
	}

	// Test no SET expressions
	query2 := Update("users")
	_, _, err = query2.SQL()
	if err == nil {
		t.Error("expected error for no SET expressions")
	}
}

func TestUpdateBuilder_WithFunctions(t *testing.T) {
	// Test UPDATE with JSON function
	query := Update(userTable.Name()).
		Set(UserName, JsonExtract(UserEmail, String("$.name"))).
		Where(UserID.Eq(1))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `name`=JSON_EXTRACT(`users`.`email`, ?) WHERE `users`.`id` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sqlStr, expectedSQL)
	}

	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
}
