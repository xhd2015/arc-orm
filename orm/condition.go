package orm

import (
	"fmt"
	"reflect"

	"github.com/xhd2015/arc-orm/field"
	"github.com/xhd2015/less-gen/strcase"
)

func (o *ORM[T, P]) ToConditions(condition *P) ([]field.Condition, error) {
	if condition == nil {
		return nil, fmt.Errorf("requires condition")
	}
	var sqlConditions []field.Condition

	rv := reflect.ValueOf(condition).Elem()
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("condition must be a struct")
	}
	t := rv.Type()
	n := rv.NumField()
	for i := 0; i < n; i++ {
		fieldV := rv.Field(i)
		field := t.Field(i)
		if field.Anonymous {
			continue
		}
		fieldName := field.Name
		colName := strcase.CamelToSnake(fieldName)

		condV := fieldV
		if fieldV.Kind() == reflect.Ptr {
			if fieldV.IsNil() {
				continue
			}
			condV = fieldV.Elem()
		}

		sqlConditions = append(sqlConditions, &rawCondition{
			sql:  fmt.Sprintf("`%s` = ?", colName),
			args: []interface{}{condV.Interface()},
		})
	}

	return sqlConditions, nil
}

func (o *ORM[T, P]) toIDCondition(id int64) (field.Condition, error) {
	if id == 0 {
		return nil, fmt.Errorf("requires id, got 0")
	}
	// Validate that the table has an 'id' field
	hasIDField := false
	for _, f := range o.table.Fields() {
		if f.Name() == "id" {
			hasIDField = true
			break
		}
	}
	if !hasIDField {
		return nil, ErrMissingIDField
	}

	idField := field.Int64Field{
		FieldName: "id",
	}

	return idField.Eq(id), nil
}

type rawCondition struct {
	sql  string
	args []interface{}
}

func (c *rawCondition) ToSQL() (string, []interface{}, error) {
	return c.sql, c.args, nil
}
