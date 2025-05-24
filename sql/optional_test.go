package sql

import (
	"testing"

	"github.com/xhd2015/arc-orm/table"
)

func TestOptional(t *testing.T) {
	userTable := table.New("users")
	UserID := userTable.Int64("id")
	UserName := userTable.String("name")
	UserEmail := userTable.String("email")

	t.Run("Optional with v=true and single condition", func(t *testing.T) {
		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(Optional(true, UserID.Gt(10)))

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users` WHERE `users`.`id` > ?"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 1 || params[0] != int64(10) {
			t.Errorf("Expected params [10], got %v", params)
		}
	})

	t.Run("Optional with v=false and single condition", func(t *testing.T) {
		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(Optional(false, UserID.Gt(10)))

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users`"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 0 {
			t.Errorf("Expected no params, got %v", params)
		}
	})

	t.Run("Optional with v=true and multiple conditions", func(t *testing.T) {
		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(Optional(true, UserID.Gt(10), UserName.Like("%John%")))

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users` WHERE (`users`.`id` > ? AND `users`.`name` LIKE ?)"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 2 || params[0] != int64(10) || params[1] != "%John%" {
			t.Errorf("Expected params [10, %%John%%], got %v", params)
		}
	})

	t.Run("Optional with v=false and multiple conditions", func(t *testing.T) {
		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(Optional(false, UserID.Gt(10), UserName.Like("%John%")))

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users`"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 0 {
			t.Errorf("Expected no params, got %v", params)
		}
	})

	t.Run("Optional with v=true and no conditions", func(t *testing.T) {
		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(Optional(true))

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users`"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 0 {
			t.Errorf("Expected no params, got %v", params)
		}
	})

	t.Run("Mixed Optional and regular conditions", func(t *testing.T) {
		includeAgeFilter := true
		includeEmailFilter := false

		query := Select(UserID, UserName).
			From(userTable.Name()).
			Where(
				UserName.Like("%test%"),                                       // Always included
				Optional(includeAgeFilter, UserID.Gt(18)),                     // Included
				Optional(includeEmailFilter, UserEmail.Like("%@example.com")), // Not included
			)

		sqlStr, params, err := query.SQL()
		if err != nil {
			t.Fatalf("Failed to generate SQL: %v", err)
		}

		expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users` WHERE `users`.`name` LIKE ? AND `users`.`id` > ?"
		if sqlStr != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
		}

		if len(params) != 2 || params[0] != "%test%" || params[1] != int64(18) {
			t.Errorf("Expected params [%%test%%, 18], got %v", params)
		}
	})
}
