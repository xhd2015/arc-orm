package sql

import (
	"testing"
	"time"
)

func TestLiteralToSQL(t *testing.T) {
	tests := []struct {
		name    string
		literal interface {
			ToSQL() (string, []interface{}, error)
		}
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name:         "String literal",
			literal:      String("hello"),
			expectedSQL:  "?",
			expectedArgs: []interface{}{"hello"},
		},
		{
			name:         "Int64 literal",
			literal:      Int64(42),
			expectedSQL:  "?",
			expectedArgs: []interface{}{int64(42)},
		},
		{
			name:         "Int32 literal",
			literal:      Int32(32),
			expectedSQL:  "?",
			expectedArgs: []interface{}{int32(32)},
		},
		{
			name:         "Float64 literal",
			literal:      Float64(3.14),
			expectedSQL:  "?",
			expectedArgs: []interface{}{float64(3.14)},
		},
		{
			name:         "Bool true literal",
			literal:      Bool(true),
			expectedSQL:  "?",
			expectedArgs: []interface{}{true},
		},
		{
			name:         "Bool false literal",
			literal:      Bool(false),
			expectedSQL:  "?",
			expectedArgs: []interface{}{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := tt.literal.ToSQL()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSQL != tt.expectedSQL {
				t.Errorf("SQL mismatch: got %q, want %q", gotSQL, tt.expectedSQL)
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

func TestTimeLiteralToSQL(t *testing.T) {
	testTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	literal := Time(testTime)

	gotSQL, gotArgs, err := literal.ToSQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotSQL != "?" {
		t.Errorf("SQL mismatch: got %q, want %q", gotSQL, "?")
	}

	if len(gotArgs) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(gotArgs))
	}

	// Check the time value
	gotTime, ok := gotArgs[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", gotArgs[0])
	}
	if !gotTime.Equal(testTime) {
		t.Errorf("time mismatch: got %v, want %v", gotTime, testTime)
	}
}

func TestLiteralsInUpdate(t *testing.T) {
	// Test that all literal types work correctly in UPDATE statements
	t.Run("String in UPDATE", func(t *testing.T) {
		query := Update("test").Set(UserName, String("value"))
		sql, args, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != "UPDATE `test` SET `name`=?" {
			t.Errorf("SQL mismatch: got %s", sql)
		}
		if len(args) != 1 || args[0] != "value" {
			t.Errorf("args mismatch: got %v", args)
		}
	})

	t.Run("Int64 in UPDATE", func(t *testing.T) {
		query := Update("test").Set(UserAge, Int64(25))
		sql, args, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != "UPDATE `test` SET `age`=?" {
			t.Errorf("SQL mismatch: got %s", sql)
		}
		if len(args) != 1 || args[0] != int64(25) {
			t.Errorf("args mismatch: got %v", args)
		}
	})

	t.Run("Bool in UPDATE", func(t *testing.T) {
		query := Update("test").Set(UserName, Bool(true))
		sql, args, err := query.SQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != "UPDATE `test` SET `name`=?" {
			t.Errorf("SQL mismatch: got %s", sql)
		}
		if len(args) != 1 || args[0] != true {
			t.Errorf("args mismatch: got %v", args)
		}
	})
}

func TestLiteralsInInsert(t *testing.T) {
	// Test that all literal types work correctly in INSERT statements
	query := InsertInto("test").
		Set(UserName, String("John")).
		Set(UserAge, Int64(30))

	sql, args, err := query.SQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "INSERT INTO `test` SET `name`=?, `age`=?"
	if sql != expectedSQL {
		t.Errorf("SQL mismatch: got %s, want %s", sql, expectedSQL)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
	if args[0] != "John" {
		t.Errorf("args[0] mismatch: got %v", args[0])
	}
	if args[1] != int64(30) {
		t.Errorf("args[1] mismatch: got %v", args[1])
	}
}
