package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql/expr"
)

// InsertInto creates a new InsertIntoBuilder for the given table
func InsertInto(tableName string) *InsertIntoBuilder {
	return &InsertIntoBuilder{
		tableName: tableName,
	}
}

// InsertIntoBuilder builds INSERT INTO queries
type InsertIntoBuilder struct {
	tableName string
	updates   []updateExpr
	err       error
}

// Set adds a column-value pair for insertion
// Value must implement expr.Expr
func (b *InsertIntoBuilder) Set(f field.Field, value expr.Expr) *InsertIntoBuilder {
	if b.err != nil {
		return b // Skip if already errored
	}
	exprSQL, exprParams, err := value.ToSQL()
	if err != nil {
		b.err = fmt.Errorf("SET field '%s': %w", f.Name(), err)
		return b
	}
	b.updates = append(b.updates, updateExpr{
		field:  f,
		expr:   exprSQL,
		params: exprParams,
	})
	return b
}

// SQL generates the SQL string and parameters
func (b *InsertIntoBuilder) SQL() (string, []interface{}, error) {
	// Check for staged errors first
	if b.err != nil {
		return "", nil, b.err
	}
	if b.tableName == "" {
		return "", nil, errors.New("table name is required")
	}
	if len(b.updates) == 0 {
		return "", nil, errors.New("no columns specified")
	}

	var sqlBuilder strings.Builder
	var params []interface{}

	// Build INSERT INTO clause
	sqlBuilder.WriteString("INSERT INTO `")
	sqlBuilder.WriteString(b.tableName)
	sqlBuilder.WriteString("` SET ")

	// Build SET clause
	for i, update := range b.updates {
		if i > 0 {
			sqlBuilder.WriteString(", ")
		}
		sqlBuilder.WriteString("`")
		sqlBuilder.WriteString(update.field.Name())
		sqlBuilder.WriteString("`=")
		sqlBuilder.WriteString(update.expr)
		params = append(params, update.params...)
	}

	return sqlBuilder.String(), params, nil
}
