package sql

import (
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/sql/expr"
)

type rand struct {
}

func (c rand) ToSQL() (string, []interface{}, error) {
	return "RAND()", nil, nil
}

func Rand() rand {
	return rand{}
}

// Func creates an arbitrary SQL function call.
// Example: Func("JSON_EXTRACT", column, sql.String("$.name")) generates JSON_EXTRACT(`table`.`column`, ?)
// Example: Func("COALESCE", column, sql.String("default")) generates COALESCE(`table`.`column`, ?)
type sqlFunc struct {
	name string
	args []expr.Expr
}

// Func creates an arbitrary SQL function call.
// All arguments must implement expr.Expr
func Func(name string, args ...expr.Expr) *sqlFunc {
	return &sqlFunc{
		name: name,
		args: args,
	}
}

// ToSQL implements expr.Expr for SQL function calls
func (f *sqlFunc) ToSQL() (string, []interface{}, error) {
	var sqlParts []string
	var params []interface{}

	for i, arg := range f.args {
		sql, p, err := arg.ToSQL()
		if err != nil {
			return "", nil, fmt.Errorf("arg[%d]: %w", i, err)
		}
		sqlParts = append(sqlParts, sql)
		params = append(params, p...)
	}

	return f.name + "(" + strings.Join(sqlParts, ", ") + ")", params, nil
}

// As returns an aliased version of this function
func (f *sqlFunc) As(alias string) *aliasedExpr {
	return &aliasedExpr{expr: f, alias: alias}
}

// Desc returns a descending order specification for this function
func (f *sqlFunc) Desc() OrderField {
	return OrderField{Field: f, Desc: true}
}

// Asc returns an ascending order specification for this function
func (f *sqlFunc) Asc() OrderField {
	return OrderField{Field: f, Desc: false}
}

// aliasedExpr wraps an expression with an alias
type aliasedExpr struct {
	expr  expr.Expr
	alias string
}

// ToSQL implements expr.Expr for aliased expressions
func (a *aliasedExpr) ToSQL() (string, []interface{}, error) {
	sql, params, err := a.expr.ToSQL()
	if err != nil {
		return "", nil, err
	}
	return sql + " AS `" + a.alias + "`", params, nil
}

// Concat creates a CONCAT SQL function call.
// Example: Concat(field, sql.String(" suffix")) generates CONCAT(`table`.`field`, ?)
func Concat(args ...expr.Expr) *sqlFunc {
	return Func("CONCAT", args...)
}

// Coalesce creates a COALESCE SQL function call.
// Example: Coalesce(field, sql.String("default")) generates COALESCE(`table`.`field`, ?)
func Coalesce(args ...expr.Expr) *sqlFunc {
	return Func("COALESCE", args...)
}

// IfNull creates an IFNULL SQL function call.
// Example: IfNull(field, sql.String("default")) generates IFNULL(`table`.`field`, ?)
func IfNull(f expr.Expr, defaultValue expr.Expr) *sqlFunc {
	return Func("IFNULL", f, defaultValue)
}

// JSON functions for convenience

// JsonExtract creates a JSON_EXTRACT SQL function call.
// Example: JsonExtract(dataField, sql.String("$.name")) generates JSON_EXTRACT(`table`.`data`, ?)
func JsonExtract(json expr.Expr, path expr.Expr) *sqlFunc {
	return Func("JSON_EXTRACT", json, path)
}

// JsonUnquote creates a JSON_UNQUOTE SQL function call.
// Example: JsonUnquote(JsonExtract(data, sql.String("$.name"))) generates JSON_UNQUOTE(JSON_EXTRACT(...))
func JsonUnquote(json expr.Expr) *sqlFunc {
	return Func("JSON_UNQUOTE", json)
}

// JsonSet creates a JSON_SET SQL function call.
// Example: JsonSet(dataField, sql.String("$.name"), sql.String("John"))
func JsonSet(json expr.Expr, pathValues ...expr.Expr) *sqlFunc {
	args := make([]expr.Expr, 0, 1+len(pathValues))
	args = append(args, json)
	args = append(args, pathValues...)
	return Func("JSON_SET", args...)
}

// JsonInsert creates a JSON_INSERT SQL function call (inserts only if path doesn't exist).
// Example: JsonInsert(dataField, sql.String("$.name"), sql.String("John"))
func JsonInsert(json expr.Expr, pathValues ...expr.Expr) *sqlFunc {
	args := make([]expr.Expr, 0, 1+len(pathValues))
	args = append(args, json)
	args = append(args, pathValues...)
	return Func("JSON_INSERT", args...)
}

// JsonReplace creates a JSON_REPLACE SQL function call (replaces only if path exists).
// Example: JsonReplace(dataField, sql.String("$.name"), sql.String("John"))
func JsonReplace(json expr.Expr, pathValues ...expr.Expr) *sqlFunc {
	args := make([]expr.Expr, 0, 1+len(pathValues))
	args = append(args, json)
	args = append(args, pathValues...)
	return Func("JSON_REPLACE", args...)
}

// JsonRemove creates a JSON_REMOVE SQL function call.
// Example: JsonRemove(dataField, sql.String("$.name"))
func JsonRemove(json expr.Expr, paths ...expr.Expr) *sqlFunc {
	args := make([]expr.Expr, 0, 1+len(paths))
	args = append(args, json)
	args = append(args, paths...)
	return Func("JSON_REMOVE", args...)
}

// JsonArray creates a JSON_ARRAY SQL function call.
// Example: JsonArray(sql.String("a"), sql.String("b")) generates JSON_ARRAY(?, ?)
func JsonArray(values ...expr.Expr) *sqlFunc {
	return Func("JSON_ARRAY", values...)
}

// JsonObject creates a JSON_OBJECT SQL function call.
// Example: JsonObject(sql.String("name"), sql.String("John")) generates JSON_OBJECT(?, ?)
func JsonObject(keyValues ...expr.Expr) *sqlFunc {
	return Func("JSON_OBJECT", keyValues...)
}

// JsonContains creates a JSON_CONTAINS SQL function call.
// Example: JsonContains(dataField, sql.String(`"value"`)) generates JSON_CONTAINS(`data`, ?)
func JsonContains(json expr.Expr, value expr.Expr) *sqlFunc {
	return Func("JSON_CONTAINS", json, value)
}

// JsonLength creates a JSON_LENGTH SQL function call.
// Example: JsonLength(dataField) generates JSON_LENGTH(`table`.`data`)
func JsonLength(json expr.Expr) *sqlFunc {
	return Func("JSON_LENGTH", json)
}

// JsonType creates a JSON_TYPE SQL function call.
// Example: JsonType(dataField) generates JSON_TYPE(`table`.`data`)
func JsonType(json expr.Expr) *sqlFunc {
	return Func("JSON_TYPE", json)
}

// JsonValid creates a JSON_VALID SQL function call.
// Example: JsonValid(dataField) generates JSON_VALID(`table`.`data`)
func JsonValid(json expr.Expr) *sqlFunc {
	return Func("JSON_VALID", json)
}

// JsonKeys creates a JSON_KEYS SQL function call.
// Example: JsonKeys(dataField) generates JSON_KEYS(`table`.`data`)
func JsonKeys(json expr.Expr) *sqlFunc {
	return Func("JSON_KEYS", json)
}

// JsonSearch creates a JSON_SEARCH SQL function call.
// Example: JsonSearch(dataField, sql.String("one"), sql.String("value"))
func JsonSearch(json expr.Expr, oneOrAll expr.Expr, searchStr expr.Expr) *sqlFunc {
	return Func("JSON_SEARCH", json, oneOrAll, searchStr)
}

// Date creates a DATE SQL function call to extract the date part from a datetime.
// Example: Date(createdAt) generates DATE(`table`.`created_at`)
func Date(f expr.Expr) *sqlFunc {
	return Func("DATE", f)
}
