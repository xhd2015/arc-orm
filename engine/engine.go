package engine

import (
	"context"
)

// Factory is responsible for creating an Engine
// its purpose is to separate initialization and usage
// without this factory, we need to ensure engine is
// initialized before ORM is created, which is not
// flexible and convenient
type Factory interface {
	GetEngine() Engine
}

type Engine interface {
	// Query execute a query sql, and return the result
	Query(ctx context.Context, sql string, args []interface{}, result interface{}) error

	// Exec execute a sql
	Exec(ctx context.Context, sql string, args []interface{}) error

	// ExecInsert execute an insert sql, and return the last insert id
	// it is essentially the same as Exec, but with a return value
	ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error)
}

// Getter is a function that returns an Engine
type Getter func() Engine

// Get returns the Engine
func (f Getter) GetEngine() Engine {
	return f()
}
