package sql

import (
	"testing"

	"github.com/xhd2015/ormx/example/user"
)

func TestDeleteFromBasic(t *testing.T) {
	// Test basic DELETE query
	query := DeleteFrom(user.Table.Name()).
		Where(user.ID.Eq(1))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "DELETE FROM `users` WHERE `users`.`id` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 1 {
		t.Errorf("Expected param to be int64(1), got %T %v", params[0], params[0])
	}
}

func TestDeleteFromMultipleConditions(t *testing.T) {
	// Test DELETE with multiple conditions
	query := DeleteFrom(user.Table.Name()).
		Where(
			user.Name.Like("%John%"),
			user.Age.Gt(30),
		)

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "DELETE FROM `users` WHERE `users`.`name` LIKE ? AND `users`.`age` > ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}
	if params[0] != "%John%" {
		t.Errorf("Expected first param to be '%%John%%', got %v", params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 30 {
		t.Errorf("Expected second param to be int64(30), got %T %v", params[1], params[1])
	}
}

func TestDeleteFromWithLimit(t *testing.T) {
	// Test DELETE with LIMIT
	query := DeleteFrom(user.Table.Name()).
		Where(user.Age.Lt(18)).
		Limit(10)

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "DELETE FROM `users` WHERE `users`.`age` < ? LIMIT 10"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 18 {
		t.Errorf("Expected param to be int64(18), got %T %v", params[0], params[0])
	}
}

func TestDeleteFromAllRows(t *testing.T) {
	// Test DELETE without WHERE (all rows)
	query := DeleteFrom(user.Table.Name())

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "DELETE FROM `users`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}
}
