package sql

import (
	"testing"

	"github.com/xhd2015/arc-orm/table"
)

func TestFunc(t *testing.T) {
	// Create test table
	testTable := table.New("users")
	id := testTable.Int64("id")
	name := testTable.String("name")
	data := testTable.String("data")

	tests := []struct {
		name         string
		funcCall     *sqlFunc
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name:         "simple function with field",
			funcCall:     Func("UPPER", name),
			expectedSQL:  "UPPER(`users`.`name`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JSON_EXTRACT with field and path",
			funcCall:     Func("JSON_EXTRACT", data, String("$.name")),
			expectedSQL:  "JSON_EXTRACT(`users`.`data`, ?)",
			expectedArgs: []interface{}{"$.name"},
		},
		{
			name:         "COALESCE with field and default",
			funcCall:     Coalesce(name, String("unknown")),
			expectedSQL:  "COALESCE(`users`.`name`, ?)",
			expectedArgs: []interface{}{"unknown"},
		},
		{
			name:         "IFNULL with field and default",
			funcCall:     IfNull(name, String("N/A")),
			expectedSQL:  "IFNULL(`users`.`name`, ?)",
			expectedArgs: []interface{}{"N/A"},
		},
		{
			name:         "CONCAT with multiple args",
			funcCall:     Concat(name, String(" "), String("suffix")),
			expectedSQL:  "CONCAT(`users`.`name`, ?, ?)",
			expectedArgs: []interface{}{" ", "suffix"},
		},
		{
			name:         "nested function",
			funcCall:     Func("UPPER", Func("TRIM", name)),
			expectedSQL:  "UPPER(TRIM(`users`.`name`))",
			expectedArgs: []interface{}{},
		},
		{
			name:         "function with literal int",
			funcCall:     Func("SUBSTRING", name, Int64(1), Int64(10)),
			expectedSQL:  "SUBSTRING(`users`.`name`, ?, ?)",
			expectedArgs: []interface{}{int64(1), int64(10)},
		},
		{
			name:         "no args",
			funcCall:     Func("NOW"),
			expectedSQL:  "NOW()",
			expectedArgs: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.funcCall.ToSQL()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSQL != tt.expectedSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tt.expectedSQL)
			}

			if len(gotArgs) != len(tt.expectedArgs) {
				t.Errorf("args length mismatch: got %d, want %d", len(gotArgs), len(tt.expectedArgs))
				return
			}

			for i, want := range tt.expectedArgs {
				if gotArgs[i] != want {
					t.Errorf("arg[%d] mismatch: got %v (%T), want %v (%T)",
						i, gotArgs[i], gotArgs[i], want, want)
				}
			}
		})
	}

	// Test in WHERE clause
	t.Run("func in WHERE clause", func(t *testing.T) {
		lengthFunc := Func("LENGTH", name)
		query := Select(id, name).
			From(testTable.Name()).
			Where(lengthFunc)

		sql, _, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The function should appear in the WHERE clause
		expectedContains := "WHERE LENGTH(`users`.`name`)"
		if !contains(sql, expectedContains) {
			t.Errorf("expected SQL to contain %q, got %s", expectedContains, sql)
		}
	})
}

func TestFuncInUpdate(t *testing.T) {
	// Create test table
	testTable := table.New("users")
	id := testTable.Int64("id")
	name := testTable.String("name")

	// Test UPPER function in update
	query := Update(testTable.Name()).
		Set(name, Func("UPPER", String("test"))).
		Where(id.Eq(1))

	sql, args, err := query.SQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `name`=UPPER(?) WHERE `users`.`id` = ?"
	if sql != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d: %v", len(args), args)
	} else {
		if args[0] != "test" {
			t.Errorf("arg[0] mismatch: got %v, want 'test'", args[0])
		}
	}
}

func TestFuncToSQL(t *testing.T) {
	// Create test table
	testTable := table.New("users")
	name := testTable.String("name")

	tests := []struct {
		name         string
		funcCall     *sqlFunc
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name:         "single literal param",
			funcCall:     Func("UPPER", String("test")),
			expectedSQL:  "UPPER(?)",
			expectedArgs: []interface{}{"test"},
		},
		{
			name:         "single field param",
			funcCall:     Func("UPPER", name),
			expectedSQL:  "UPPER(`users`.`name`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "field + literal",
			funcCall:     Func("CONCAT", name, String("suffix")),
			expectedSQL:  "CONCAT(`users`.`name`, ?)",
			expectedArgs: []interface{}{"suffix"},
		},
		{
			name:         "no params",
			funcCall:     Func("NOW"),
			expectedSQL:  "NOW()",
			expectedArgs: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.funcCall.ToSQL()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSQL != tt.expectedSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tt.expectedSQL)
			}

			if len(gotArgs) != len(tt.expectedArgs) {
				t.Errorf("args length mismatch: got %d, want %d", len(gotArgs), len(tt.expectedArgs))
				return
			}

			for i, want := range tt.expectedArgs {
				if gotArgs[i] != want {
					t.Errorf("arg[%d] mismatch: got %v (%T), want %v (%T)",
						i, gotArgs[i], gotArgs[i], want, want)
				}
			}
		})
	}

	// Test that UpdateBuilder works correctly with functions
	t.Run("UpdateBuilder with CONCAT", func(t *testing.T) {
		query := Update(testTable.Name()).
			Set(name, Func("CONCAT", name, String(" suffix"))).
			Where(testTable.Int64("id").Eq(1))

		sql, args, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "UPDATE `users` SET `name`=CONCAT(`users`.`name`, ?) WHERE `users`.`id` = ?"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}

		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		} else {
			if args[0] != " suffix" {
				t.Errorf("arg[0] mismatch: got %v, want ' suffix'", args[0])
			}
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDateFunc(t *testing.T) {
	// Create test table
	testTable := table.New("tasks")
	createdAt := testTable.Time("created_at")

	t.Run("DATE function", func(t *testing.T) {
		dateFunc := Date(createdAt)
		sql, args, err := dateFunc.ToSQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "DATE(`tasks`.`created_at`)"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}

		if len(args) != 0 {
			t.Errorf("expected 0 args, got %d", len(args))
		}
	})

	t.Run("DATE in SELECT", func(t *testing.T) {
		query := Select(Date(createdAt)).From(testTable.Name())
		sql, _, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "SELECT DATE(`tasks`.`created_at`) FROM `tasks`"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}
	})

	t.Run("DATE in GROUP BY", func(t *testing.T) {
		query := Select(Date(createdAt), Count(All).As("count")).
			From(testTable.Name()).
			GroupBy(Date(createdAt))
		sql, _, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "SELECT DATE(`tasks`.`created_at`), COUNT(*) AS `count` FROM `tasks` GROUP BY DATE(`tasks`.`created_at`)"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}
	})
}

func TestFuncAs(t *testing.T) {
	// Create test table
	testTable := table.New("tasks")
	createdAt := testTable.Time("created_at")

	t.Run("As alias", func(t *testing.T) {
		aliased := Date(createdAt).As("date")
		sql, args, err := aliased.ToSQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "DATE(`tasks`.`created_at`) AS `date`"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}

		if len(args) != 0 {
			t.Errorf("expected 0 args, got %d", len(args))
		}
	})

	t.Run("As in SELECT", func(t *testing.T) {
		query := Select(Date(createdAt).As("date"), Count(All).As("count")).
			From(testTable.Name()).
			GroupBy(Date(createdAt))
		sql, _, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "SELECT DATE(`tasks`.`created_at`) AS `date`, COUNT(*) AS `count` FROM `tasks` GROUP BY DATE(`tasks`.`created_at`)"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}
	})
}

func TestFuncDescAsc(t *testing.T) {
	// Create test table
	testTable := table.New("tasks")
	createdAt := testTable.Time("created_at")

	t.Run("Desc ordering", func(t *testing.T) {
		orderField := Date(createdAt).Desc()
		sql, args, err := orderField.ToSQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "DATE(`tasks`.`created_at`) DESC"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}

		if len(args) != 0 {
			t.Errorf("expected 0 args, got %d", len(args))
		}
	})

	t.Run("Asc ordering", func(t *testing.T) {
		orderField := Date(createdAt).Asc()
		sql, args, err := orderField.ToSQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "DATE(`tasks`.`created_at`) ASC"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}

		if len(args) != 0 {
			t.Errorf("expected 0 args, got %d", len(args))
		}
	})

	t.Run("ORDER BY with Desc", func(t *testing.T) {
		query := Select(Date(createdAt).As("date"), Count(All).As("count")).
			From(testTable.Name()).
			GroupBy(Date(createdAt)).
			OrderBy(Date(createdAt).Desc())
		sql, _, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedSQL := "SELECT DATE(`tasks`.`created_at`) AS `date`, COUNT(*) AS `count` FROM `tasks` GROUP BY DATE(`tasks`.`created_at`) ORDER BY DATE(`tasks`.`created_at`) DESC"
		if sql != expectedSQL {
			t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", sql, expectedSQL)
		}
	})
}

func TestJsonFuncs(t *testing.T) {
	// Create test table
	testTable := table.New("users")
	data := testTable.String("data")

	tests := []struct {
		name         string
		funcCall     *sqlFunc
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name:         "JsonExtract",
			funcCall:     JsonExtract(data, String("$.name")),
			expectedSQL:  "JSON_EXTRACT(`users`.`data`, ?)",
			expectedArgs: []interface{}{"$.name"},
		},
		{
			name:         "JsonUnquote",
			funcCall:     JsonUnquote(data),
			expectedSQL:  "JSON_UNQUOTE(`users`.`data`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JsonUnquote with JsonExtract",
			funcCall:     JsonUnquote(JsonExtract(data, String("$.name"))),
			expectedSQL:  "JSON_UNQUOTE(JSON_EXTRACT(`users`.`data`, ?))",
			expectedArgs: []interface{}{"$.name"},
		},
		{
			name:         "JsonSet",
			funcCall:     JsonSet(data, String("$.name"), String("John")),
			expectedSQL:  "JSON_SET(`users`.`data`, ?, ?)",
			expectedArgs: []interface{}{"$.name", "John"},
		},
		{
			name:         "JsonInsert",
			funcCall:     JsonInsert(data, String("$.age"), Int64(25)),
			expectedSQL:  "JSON_INSERT(`users`.`data`, ?, ?)",
			expectedArgs: []interface{}{"$.age", int64(25)},
		},
		{
			name:         "JsonReplace",
			funcCall:     JsonReplace(data, String("$.name"), String("Jane")),
			expectedSQL:  "JSON_REPLACE(`users`.`data`, ?, ?)",
			expectedArgs: []interface{}{"$.name", "Jane"},
		},
		{
			name:         "JsonRemove",
			funcCall:     JsonRemove(data, String("$.temp")),
			expectedSQL:  "JSON_REMOVE(`users`.`data`, ?)",
			expectedArgs: []interface{}{"$.temp"},
		},
		{
			name:         "JsonArray",
			funcCall:     JsonArray(String("a"), String("b"), String("c")),
			expectedSQL:  "JSON_ARRAY(?, ?, ?)",
			expectedArgs: []interface{}{"a", "b", "c"},
		},
		{
			name:         "JsonObject",
			funcCall:     JsonObject(String("name"), String("John"), String("age"), Int64(25)),
			expectedSQL:  "JSON_OBJECT(?, ?, ?, ?)",
			expectedArgs: []interface{}{"name", "John", "age", int64(25)},
		},
		{
			name:         "JsonContains",
			funcCall:     JsonContains(data, String(`"value"`)),
			expectedSQL:  "JSON_CONTAINS(`users`.`data`, ?)",
			expectedArgs: []interface{}{`"value"`},
		},
		{
			name:         "JsonLength",
			funcCall:     JsonLength(data),
			expectedSQL:  "JSON_LENGTH(`users`.`data`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JsonType",
			funcCall:     JsonType(data),
			expectedSQL:  "JSON_TYPE(`users`.`data`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JsonValid",
			funcCall:     JsonValid(data),
			expectedSQL:  "JSON_VALID(`users`.`data`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JsonKeys",
			funcCall:     JsonKeys(data),
			expectedSQL:  "JSON_KEYS(`users`.`data`)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "JsonSearch",
			funcCall:     JsonSearch(data, String("one"), String("searchValue")),
			expectedSQL:  "JSON_SEARCH(`users`.`data`, ?, ?)",
			expectedArgs: []interface{}{"one", "searchValue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.funcCall.ToSQL()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSQL != tt.expectedSQL {
				t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, tt.expectedSQL)
			}

			if len(gotArgs) != len(tt.expectedArgs) {
				t.Errorf("args length mismatch: got %d, want %d", len(gotArgs), len(tt.expectedArgs))
				return
			}

			for i, want := range tt.expectedArgs {
				if gotArgs[i] != want {
					t.Errorf("arg[%d] mismatch: got %v (%T), want %v (%T)",
						i, gotArgs[i], gotArgs[i], want, want)
				}
			}
		})
	}
}
