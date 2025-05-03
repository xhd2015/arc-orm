package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/field"
)

// DeleteFrom creates a new DeleteBuilder for the given table
func DeleteFrom(tableName string) *DeleteBuilder {
	return &DeleteBuilder{
		tableName: tableName,
	}
}

// DeleteBuilder builds DELETE queries
type DeleteBuilder struct {
	tableName  string
	conditions []field.Condition
	limit      int
	hasLimit   bool
}

// Where adds conditions to the DELETE query
func (b *DeleteBuilder) Where(conditions ...field.Condition) *DeleteBuilder {
	b.conditions = append(b.conditions, conditions...)
	return b
}

// Limit sets the maximum number of rows to delete
func (b *DeleteBuilder) Limit(limit int) *DeleteBuilder {
	b.limit = limit
	b.hasLimit = true
	return b
}

// SQL generates the SQL string and parameters for the DELETE statement
func (b *DeleteBuilder) SQL() (string, []interface{}, error) {
	if b.tableName == "" {
		return "", nil, errors.New("table name is required")
	}

	var sqlBuilder strings.Builder
	var params []interface{}

	// Build DELETE clause
	sqlBuilder.WriteString("DELETE FROM `")
	sqlBuilder.WriteString(b.tableName)
	sqlBuilder.WriteString("`")

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

	// Add LIMIT clause if specified
	if b.hasLimit {
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}

	return sqlBuilder.String(), params, nil
}
