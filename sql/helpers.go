package sql

import (
	"strings"
)

// optimizeSQL removes table name prefixes from SQL when they match the table being operated on
func optimizeSQL(tableName string, sql string) string {
	if tableName == "" {
		return sql
	}

	// Replace table name prefix with just backtick for field names
	tablePrefix := "`" + tableName + "`."
	return strings.ReplaceAll(sql, tablePrefix, "`")
}
