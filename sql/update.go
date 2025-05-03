package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/field"
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
	conditions []field.Condition
}

// updateExpr represents an update expression in the SET clause
type updateExpr struct {
	field field.Field
	expr  string
	value interface{}
}

// Set adds a field=value expression to the SET clause
// Value must be a field.Expression
func (b *UpdateBuilder) Set(f field.Field, value field.Expression) *UpdateBuilder {
	exprSQL, exprValue := value.ToExpressionSQL()
	b.updates = append(b.updates, updateExpr{
		field: f,
		expr:  "=" + exprSQL,
		value: exprValue,
	})
	return b
}

// Where adds conditions to the UPDATE query
func (b *UpdateBuilder) Where(conditions ...field.Condition) *UpdateBuilder {
	b.conditions = append(b.conditions, conditions...)
	return b
}

// SQL generates the SQL string and parameters for the UPDATE statement
func (b *UpdateBuilder) SQL() (string, []interface{}, error) {
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
		sqlBuilder.WriteString("?")
		params = append(params, update.value)
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

			sqlBuilder.WriteString(condSQL)
			params = append(params, condParams...)
		}
	}

	return sqlBuilder.String(), params, nil
}
