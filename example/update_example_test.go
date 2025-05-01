package example

import (
	"fmt"
	"testing"

	"github.com/xhd2015/ormx/example/user"
	"github.com/xhd2015/ormx/sql"
)

func TestUpdateExamples(t *testing.T) {
	// Example 1: Basic update with literal values
	query1 := sql.Update(user.Table.Name()).
		Set(user.Name, sql.String("John Doe")).
		Set(user.Age, sql.Int64(30)).
		Where(user.ID.Eq(1))

	sql1, params1, _ := query1.SQL()
	fmt.Printf("Example 1: %s with params %v\n", sql1, params1)

	// Example 2: Using field expressions
	query2 := sql.Update(user.Table.Name()).
		Set(user.Age, user.Age.Increment(1)).
		Set(user.Email, user.Email.Concat("@example.com")).
		Where(user.ID.Eq(2))

	sql2, params2, _ := query2.SQL()
	fmt.Printf("Example 2: %s with params %v\n", sql2, params2)

	// Example 3: Different literal types
	query3 := sql.Update(user.Table.Name()).
		Set(user.Name, sql.String("Jane Doe")).
		Set(user.Email, sql.String("jane@example.com")).
		Where(user.ID.Eq(3))

	sql3, params3, _ := query3.SQL()
	fmt.Printf("Example 3: %s with params %v\n", sql3, params3)
}
