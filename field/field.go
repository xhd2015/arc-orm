package field

import "strings"

// Field interface represents a database field with a name and table
type Field interface {
	Name() string
	Table() string
	ToSQL() string
}

// Condition represents a SQL condition
type Condition interface {
	ToSQL() (string, []interface{}, error)
}

// AliasField wraps any field with an alias
type AliasField struct {
	field Field
	alias string
}

// Name returns the field name
func (a AliasField) Name() string {
	return a.field.Name()
}

// Table returns the table name
func (a AliasField) Table() string {
	return a.field.Table()
}

// ToSQL returns the SQL representation of the field with its alias
func (a AliasField) ToSQL() string {
	return a.field.ToSQL() + " AS `" + a.alias + "`"
}

// As creates an aliased field
func As(f Field, alias string) Field {
	return AliasField{
		field: f,
		alias: alias,
	}
}

// comparison represents a comparison operation between a field and a value
type comparison struct {
	field Field
	op    string
	value interface{}
}

func (c *comparison) ToSQL() (string, []interface{}, error) {
	return c.field.ToSQL() + " " + c.op + " ?", []interface{}{c.value}, nil
}

// like represents a LIKE condition
type like struct {
	field Field
	value string
}

func (l *like) ToSQL() (string, []interface{}, error) {
	return l.field.ToSQL() + " LIKE ?", []interface{}{l.value}, nil
}

// or represents an OR condition
type or struct {
	conditions []Condition
}

type and struct {
	conditions []Condition
}

// Or creates an OR condition from multiple conditions
func Or(conditions ...Condition) Condition {
	return &or{conditions: conditions}
}
func And(conditions ...Condition) Condition {
	return &and{conditions: conditions}
}

func (o *or) ToSQL() (string, []interface{}, error) {
	return joinCodnitions(o.conditions, "OR")
}

func (a *and) ToSQL() (string, []interface{}, error) {
	return joinCodnitions(a.conditions, "AND")
}

func joinCodnitions(conditions []Condition, op string) (string, []interface{}, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	sqlParts := make([]string, 0, len(conditions))
	params := make([]interface{}, 0)

	for _, cond := range conditions {
		sql, condParams, err := cond.ToSQL()
		if err != nil {
			return "", nil, err
		}
		if sql == "" {
			continue
		}
		sqlParts = append(sqlParts, sql)
		params = append(params, condParams...)
	}
	if len(sqlParts) == 0 {
		return "", nil, nil
	}
	if len(sqlParts) == 1 {
		return sqlParts[0], params, nil
	}

	return "(" + strings.Join(sqlParts, " "+op+" ") + ")", params, nil
}

// Expression represents a SQL expression
type Expression interface {
	ToExpressionSQL() (string, interface{})
}

// fieldOperation represents a field operation (like increment, decrement, concatenate)
type fieldOperation struct {
	field    Field
	operator string
	value    interface{}
}

func (op *fieldOperation) ToExpressionSQL() (string, interface{}) {
	return "`" + op.field.Table() + "`.`" + op.field.Name() + "`" + op.operator, op.value
}
