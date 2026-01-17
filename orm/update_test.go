package orm

import (
	"context"
	"testing"

	"github.com/xhd2015/arc-orm/engine"
	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql"
	"github.com/xhd2015/arc-orm/table"
)

func TestORMUpdateBuilder_Increment(t *testing.T) {
	// Create test table with age field
	testTable := table.New("users")
	id := testTable.Int64("id")
	name := testTable.String("name")
	age := testTable.Int64("age")

	tests := []struct {
		name         string
		setup        func() *sql.UpdateBuilder
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "simple increment age=age+1",
			setup: func() *sql.UpdateBuilder {
				return sql.Update(testTable.Name()).
					Set(age, age.Increment(1)).
					Where(id.Eq(1))
			},
			expectedSQL:  "UPDATE `users` SET `age`=`users`.`age`+? WHERE `users`.`id` = ?",
			expectedArgs: []interface{}{int64(1), int64(1)},
		},
		{
			name: "increment with larger value age=age+10",
			setup: func() *sql.UpdateBuilder {
				return sql.Update(testTable.Name()).
					Set(age, age.Increment(10)).
					Where(id.Eq(2))
			},
			expectedSQL:  "UPDATE `users` SET `age`=`users`.`age`+? WHERE `users`.`id` = ?",
			expectedArgs: []interface{}{int64(10), int64(2)},
		},
		{
			name: "decrement age=age-5",
			setup: func() *sql.UpdateBuilder {
				return sql.Update(testTable.Name()).
					Set(age, age.Decrement(5)).
					Where(id.Eq(3))
			},
			expectedSQL:  "UPDATE `users` SET `age`=`users`.`age`-? WHERE `users`.`id` = ?",
			expectedArgs: []interface{}{int64(5), int64(3)},
		},
		{
			name: "increment with other field update",
			setup: func() *sql.UpdateBuilder {
				return sql.Update(testTable.Name()).
					Set(name, sql.String("John")).
					Set(age, age.Increment(1)).
					Where(id.Eq(4))
			},
			expectedSQL:  "UPDATE `users` SET `name`=?, `age`=`users`.`age`+? WHERE `users`.`id` = ?",
			expectedArgs: []interface{}{"John", int64(1), int64(4)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			gotSQL, gotArgs, err := builder.SQL()
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

func TestORMUpdateBuilder_IncrementWithORM(t *testing.T) {
	// Mock engine that captures the executed SQL
	var capturedSQL string
	var capturedArgs []interface{}
	mockEngine := &MockExecEngine{
		ExecFunc: func(ctx context.Context, sql string, args []interface{}) error {
			capturedSQL = sql
			capturedArgs = args
			return nil
		},
	}

	// Create test table
	testTable := table.New("users")
	id := testTable.Int64("id")
	testTable.String("name")
	age := testTable.Int64("age")

	// Create ORM instance
	orm := &ORM[TestModel, TestModelOptional]{
		table:  testTable,
		engine: mockEngine,
	}

	// Execute update with increment
	err := orm.Update().
		Set(age, age.Increment(1)).
		Where(id.Eq(1)).
		Exec(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `age`=`users`.`age`+? WHERE `users`.`id` = ?"
	if capturedSQL != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", capturedSQL, expectedSQL)
	}

	expectedArgs := []interface{}{int64(1), int64(1)}
	if len(capturedArgs) != len(expectedArgs) {
		t.Errorf("args length mismatch: got %d, want %d", len(capturedArgs), len(expectedArgs))
		return
	}

	for i, want := range expectedArgs {
		if capturedArgs[i] != want {
			t.Errorf("arg[%d] mismatch: got %v (%T), want %v (%T)",
				i, capturedArgs[i], capturedArgs[i], want, want)
		}
	}
}

func TestORMUpdateBuilder_Float64Increment(t *testing.T) {
	// Create test table with float field
	testTable := table.New("products")
	id := testTable.Int64("id")
	price := testTable.Float64("price")

	// Test float64 increment
	builder := sql.Update(testTable.Name()).
		Set(price, price.Increment(0.5)).
		Where(id.Eq(1))

	gotSQL, gotArgs, err := builder.SQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `products` SET `price`=`products`.`price`+? WHERE `products`.`id` = ?"
	if gotSQL != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, expectedSQL)
	}

	if len(gotArgs) != 2 {
		t.Errorf("expected 2 args, got %d", len(gotArgs))
		return
	}

	if gotArgs[0] != 0.5 {
		t.Errorf("arg[0] mismatch: got %v, want 0.5", gotArgs[0])
	}
}

func TestORMUpdateBuilder_Int32Increment(t *testing.T) {
	// Create test table with int32 field
	testTable := table.New("items")
	id := testTable.Int64("id")
	quantity := testTable.Int32("quantity")

	// Test int32 increment
	builder := sql.Update(testTable.Name()).
		Set(quantity, quantity.Increment(5)).
		Where(id.Eq(1))

	gotSQL, gotArgs, err := builder.SQL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `items` SET `quantity`=`items`.`quantity`+? WHERE `items`.`id` = ?"
	if gotSQL != expectedSQL {
		t.Errorf("SQL mismatch:\n  got:  %s\n  want: %s", gotSQL, expectedSQL)
	}

	if len(gotArgs) != 2 {
		t.Errorf("expected 2 args, got %d", len(gotArgs))
		return
	}

	if gotArgs[0] != int32(5) {
		t.Errorf("arg[0] mismatch: got %v (%T), want int32(5)", gotArgs[0], gotArgs[0])
	}
}

// MockExecEngine for testing - implements engine.Factory and engine.Engine
type MockExecEngine struct {
	ExecFunc func(ctx context.Context, sql string, args []interface{}) error
}

func (m *MockExecEngine) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	return nil
}

func (m *MockExecEngine) Exec(ctx context.Context, sql string, args []interface{}) error {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args)
	}
	return nil
}

func (m *MockExecEngine) ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error) {
	return 0, nil
}

func (m *MockExecEngine) GetEngine() engine.Engine {
	return m
}

// testFieldExpression implements field.Expr for testing
type testFieldExpression struct {
	sql    string
	params []interface{}
}

func (f testFieldExpression) ToSQL() (string, []interface{}, error) {
	return f.sql, f.params, nil
}

var _ field.Expr = testFieldExpression{}
