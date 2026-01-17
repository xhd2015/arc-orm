package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/arc-orm/sql/expr"
)

// Update creates a new UpdateBuilder for the given table
func Update(tableName string) *UpdateBuilder {
	return &UpdateBuilder{
		tableName: tableName,
	}
}

// UpdateBuilder builds UPDATE queries
type UpdateBuilder struct {
	tableName  string
	updates    []updateExpr
	conditions []expr.Expr
	err        error
}

// updateExpr represents an update expression in the SET clause
type updateExpr struct {
	field  field.Field
	expr   string
	params []interface{}
}

// Set adds a field=value expression to the SET clause
// Value must implement expr.Expr
func (b *UpdateBuilder) Set(f field.Field, value expr.Expr) *UpdateBuilder {
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
		expr:   "=" + exprSQL,
		params: exprParams,
	})
	return b
}

// Where adds conditions to the UPDATE query
func (b *UpdateBuilder) Where(conditions ...expr.Expr) *UpdateBuilder {
	if b.err != nil {
		return b // Skip if already errored
	}
	b.conditions = append(b.conditions, conditions...)
	return b
}

// SQL generates the SQL string and parameters for the UPDATE statement
func (b *UpdateBuilder) SQL() (string, []interface{}, error) {
	// Check for staged errors first
	if b.err != nil {
		return "", nil, b.err
	}
	if b.tableName == "" {
		return "", nil, errors.New("table name is required")
	}
	if len(b.updates) == 0 {
		return "", nil, errors.New("at least one SET expression is required")
	}

	var sqlBuilder strings.Builder
	var params []interface{}

	// Build UPDATE clause
	sqlBuilder.WriteString("UPDATE `")
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
		sqlBuilder.WriteString(update.expr)
		params = append(params, update.params...)
	}

	// Build WHERE clause
	if len(b.conditions) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		for i, condition := range b.conditions {
			if i > 0 {
				sqlBuilder.WriteString(" AND ")
			}

			condSQL, condParams, err := condition.ToSQL()
			if err != nil {
				return "", nil, fmt.Errorf("failed to build where condition: %w", err)
			}
			if condSQL == "" {
				continue
			}

			sqlBuilder.WriteString(condSQL)
			params = append(params, condParams...)
		}
	}

	return sqlBuilder.String(), params, nil
}
