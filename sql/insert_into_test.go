package sql

import (
	"testing"
	"time"
)

func TestInsertIntoBasic(t *testing.T) {
	// Test basic INSERT query
	query := InsertInto(userTable.Name()).
		Set(UserName, String("John Doe")).
		Set(UserEmail, String("john@example.com")).
		Set(UserAge, Int64(30))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "INSERT INTO `users` SET `name`=?, `email`=?, `age`=?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}
	if params[0] != "John Doe" {
		t.Errorf("Expected first param to be 'John Doe', got %v", params[0])
	}
	if params[1] != "john@example.com" {
		t.Errorf("Expected second param to be 'john@example.com', got %v", params[1])
	}
	if v, ok := params[2].(int64); !ok || v != 30 {
		t.Errorf("Expected third param to be int64(30), got %T %v", params[2], params[2])
	}
}

func TestInsertIntoWithTimeExpression(t *testing.T) {
	// Test INSERT with expressions
	now := time.Now()
	query := InsertInto(userTable.Name()).
		Set(UserName, String("Jane Smith")).
		Set(UserEmail, String("jane@example.com")).
		Set(UserCreateTime, Time(now))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "INSERT INTO `users` SET `name`=?, `email`=?, `create_time`=?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}
	if params[0] != "Jane Smith" {
		t.Errorf("Expected first param to be 'Jane Smith', got %v", params[0])
	}
	if params[1] != "jane@example.com" {
		t.Errorf("Expected second param to be 'jane@example.com', got %v", params[1])
	}
	if params[2] != now {
		t.Errorf("Expected third param to be time %v, got %v", now, params[2])
	}
}
