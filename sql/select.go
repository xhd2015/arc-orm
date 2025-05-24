package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/arc-orm/field"
)

// Select creates a new SelectBuilder with the given fields
func Select(fields ...field.Field) *SelectBuilder {
	return &SelectBuilder{
		fields: fields,
	}
}

// Count creates a count expression
func Count(f field.Field) AggregateFunc {
	return AggregateFunc{
		name:  "COUNT",
		field: f,
	}
}

// AggregateFunc represents an aggregate function like COUNT, SUM, etc.
type AggregateFunc struct {
	name  string
	field field.Field
}

// OrderField represents a field with ordering direction
type OrderField struct {
	Field interface{ ToSQL() string }
	Desc  bool
}

// ToSQL returns the SQL representation of the OrderField
func (o OrderField) ToSQL() string {
	sql := o.Field.ToSQL()
	if o.Desc {
		sql += " DESC"
	} else {
		sql += " ASC"
	}
	return sql
}

// ToSQL returns the SQL representation of the aggregate function
func (a AggregateFunc) ToSQL() string {
	return a.name + "(" + a.field.ToSQL() + ")"
}

// Name implements the Field interface
func (a AggregateFunc) Name() string {
	return a.name + "(" + a.field.Name() + ")"
}

// Table implements the Field interface
func (a AggregateFunc) Table() string {
	return a.field.Table()
}

// As returns this function with an alias
func (a AggregateFunc) As(alias string) field.Field {
	return field.As(a, alias)
}

// Asc returns an ascending order specification for this aggregate
func (a AggregateFunc) Asc() OrderField {
	return OrderField{Field: a, Desc: false}
}

// Desc returns a descending order specification for this aggregate
func (a AggregateFunc) Desc() OrderField {
	return OrderField{Field: a, Desc: true}
}

// Max creates a MAX expression
func Max(f field.Field) AggregateFunc {
	return AggregateFunc{
		name:  "MAX",
		field: f,
	}
}

// Gt creates a greater than condition
func (a AggregateFunc) Gt(value int64) field.Condition {
	return &havingCondition{
		expr:  a,
		op:    ">",
		value: value,
	}
}

// Lt creates a less than condition
func (a AggregateFunc) Lt(value int64) field.Condition {
	return &havingCondition{
		expr:  a,
		op:    "<",
		value: value,
	}
}

// havingCondition represents a HAVING condition
type havingCondition struct {
	expr  AggregateFunc
	op    string
	value interface{}
}

func (c *havingCondition) ToSQL() (string, []interface{}, error) {
	return c.expr.ToSQL() + " " + c.op + " ?", []interface{}{c.value}, nil
}

// SelectBuilder builds SELECT queries
type SelectBuilder struct {
	fields     []field.Field
	tableName  string
	joins      []join
	conditions []field.Condition
	groupBys   []field.Field
	havings    []field.Condition
	orderBys   []orderBy
	limit      int
	offset     int
	hasLimit   bool
	hasOffset  bool
}

type join struct {
	tableName string
	condition field.Condition
	joinType  string
}

type orderBy struct {
	field field.OrderField
}

// From specifies the table to select from
func (b *SelectBuilder) From(tableName string) *SelectBuilder {
	b.tableName = tableName
	return b
}

// Where adds conditions to the query
func (b *SelectBuilder) Where(conditions ...field.Condition) *SelectBuilder {
	b.conditions = append(b.conditions, conditions...)
	return b
}

// Join adds a join clause to the query
func (b *SelectBuilder) Join(tableName string, condition field.Condition) *SelectBuilder {
	b.joins = append(b.joins, join{
		tableName: tableName,
		condition: condition,
		joinType:  "JOIN",
	})
	return b
}

// LeftJoin adds a left join clause to the query
func (b *SelectBuilder) LeftJoin(tableName string, condition field.Condition) *SelectBuilder {
	b.joins = append(b.joins, join{
		tableName: tableName,
		condition: condition,
		joinType:  "LEFT JOIN",
	})
	return b
}

// GroupBy adds GROUP BY fields to the query
func (b *SelectBuilder) GroupBy(fields ...field.Field) *SelectBuilder {
	b.groupBys = append(b.groupBys, fields...)
	return b
}

// Having adds HAVING conditions to the query
func (b *SelectBuilder) Having(conditions ...field.Condition) *SelectBuilder {
	b.havings = append(b.havings, conditions...)
	return b
}

// OrderBy adds ORDER BY fields to the query
func (b *SelectBuilder) OrderBy(orderFields ...field.OrderField) *SelectBuilder {
	for _, f := range orderFields {
		b.orderBys = append(b.orderBys, orderBy{field: f})
	}
	return b
}

// Limit sets the LIMIT value
func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
	b.limit = limit
	b.hasLimit = true
	return b
}

// Offset sets the OFFSET value
func (b *SelectBuilder) Offset(offset int) *SelectBuilder {
	b.offset = offset
	b.hasOffset = true
	return b
}

// SQL generates the SQL string and parameters
func (b *SelectBuilder) SQL() (string, []interface{}, error) {
	if b.tableName == "" {
		return "", nil, errors.New("from table is required")
	}

	var sqlBuilder strings.Builder
	var params []interface{}

	// Build SELECT clause
	sqlBuilder.WriteString("SELECT ")
	for i, field := range b.fields {
		if i > 0 {
			sqlBuilder.WriteString(", ")
		}
		sqlBuilder.WriteString(field.ToSQL())
	}

	// Build FROM clause
	sqlBuilder.WriteString(" FROM `")
	sqlBuilder.WriteString(b.tableName)
	sqlBuilder.WriteString("`")

	// Build JOIN clauses
	for _, join := range b.joins {
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(join.joinType)
		sqlBuilder.WriteString(" `")
		sqlBuilder.WriteString(join.tableName)
		sqlBuilder.WriteString("` ON ")

		joinSQL, joinParams, err := join.condition.ToSQL()
		if err != nil {
			return "", nil, fmt.Errorf("failed to build join condition: %w", err)
		}
		if joinSQL == "" {
			continue
		}
		sqlBuilder.WriteString(joinSQL)
		params = append(params, joinParams...)
	}

	// Build WHERE clause
	if len(b.conditions) > 0 {
		whereClauses := make([]string, 0, len(b.conditions))
		for _, condition := range b.conditions {
			condSQL, condParams, err := condition.ToSQL()
			if err != nil {
				return "", nil, fmt.Errorf("failed to build where condition: %w", err)
			}
			if condSQL == "" {
				continue
			}
			whereClauses = append(whereClauses, condSQL)
			params = append(params, condParams...)
		}
		if len(whereClauses) > 0 {
			sqlBuilder.WriteString(" WHERE ")
			sqlBuilder.WriteString(strings.Join(whereClauses, " AND "))
		}
	}

	// Build GROUP BY clause
	if len(b.groupBys) > 0 {
		sqlBuilder.WriteString(" GROUP BY ")
		for i, field := range b.groupBys {
			if i > 0 {
				sqlBuilder.WriteString(", ")
			}
			sqlBuilder.WriteString(field.ToSQL())
		}
	}

	// Build HAVING clause
	if len(b.havings) > 0 {
		havingClauses := make([]string, 0, len(b.havings))
		for _, condition := range b.havings {
			condSQL, condParams, err := condition.ToSQL()
			if err != nil {
				return "", nil, fmt.Errorf("failed to build having condition: %w", err)
			}
			if condSQL == "" {
				continue
			}
			havingClauses = append(havingClauses, condSQL)
			params = append(params, condParams...)
		}
		if len(havingClauses) > 0 {
			sqlBuilder.WriteString(" HAVING ")
			sqlBuilder.WriteString(strings.Join(havingClauses, " AND "))
		}
	}

	// Build ORDER BY clause
	if len(b.orderBys) > 0 {
		sqlBuilder.WriteString(" ORDER BY ")
		for i, orderBy := range b.orderBys {
			if i > 0 {
				sqlBuilder.WriteString(", ")
			}
			sqlBuilder.WriteString(orderBy.field.ToSQL())
		}
	}

	// Add LIMIT and OFFSET
	if b.hasLimit && b.hasOffset {
		// short form
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d,%d", b.offset, b.limit))
	} else if b.hasLimit {
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	} else if b.hasOffset {
		sqlBuilder.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}

	return sqlBuilder.String(), params, nil
}
