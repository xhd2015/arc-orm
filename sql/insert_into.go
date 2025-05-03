package sql

import (
	"errors"
	"strings"

	"github.com/xhd2015/arc-orm/field"
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
}

// Set adds a column-value pair for insertion
// Value must be a field.Expression
func (b *InsertIntoBuilder) Set(f field.Field, value field.Expression) *InsertIntoBuilder {
	exprSQL, exprValue := value.ToExpressionSQL()
	b.updates = append(b.updates, updateExpr{
		field: f,
		expr:  exprSQL,
		value: exprValue,
	})
	return b
}

// SQL generates the SQL string and parameters
func (b *InsertIntoBuilder) SQL() (string, []interface{}, error) {
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
		sqlBuilder.WriteString("`")

		// If expr is empty, use "=?"
		if update.expr == "" {
			sqlBuilder.WriteString("=?")
		} else {
			sqlBuilder.WriteString("=" + update.expr)
		}

		params = append(params, update.value)
	}

	return sqlBuilder.String(), params, nil
}
