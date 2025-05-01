package sql

import (
	"strings"
	"testing"

	"github.com/xhd2015/ormx/example/comment"
	"github.com/xhd2015/ormx/example/post"
	"github.com/xhd2015/ormx/example/user"
	"github.com/xhd2015/ormx/field"
)

func TestTypeBasedSqlBuilder(t *testing.T) {
	// Basic SELECT
	query := Select(user.ID, user.Name, user.Email).
		From(user.Table.Name()).
		Where(user.ID.Eq(1))

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "SELECT `users`.`id`, `users`.`name`, `users`.`email` FROM `users` WHERE `users`.`id` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 1 {
		t.Errorf("Expected param to be int64(1), got %T %v", params[0], params[0])
	}
}

func TestJoinQueries(t *testing.T) {
	// Basic JOIN
	basicJoin := Select(
		user.ID, user.Name,
		post.ID, post.Title,
	).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID))

	sqlStr, params, err := basicJoin.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "SELECT `users`.`id`, `users`.`name`, `posts`.`id`, `posts`.`title` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}

	// More complex JOIN
	complexJoin := Select(
		user.ID, user.Name,
		post.ID, post.Title,
		comment.ID, comment.Content,
	).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID)).
		LeftJoin(comment.Table.Name(), post.ID.EqField(comment.PostID)).
		Where(
			user.ID.Gt(10),
			post.Title.Like("%golang%"),
		)

	sqlStr, params, err = complexJoin.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL = "SELECT `users`.`id`, `users`.`name`, `posts`.`id`, `posts`.`title`, `comments`.`id`, `comments`.`content` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` LEFT JOIN `comments` ON `posts`.`id` = `comments`.`post_id` WHERE `users`.`id` > ? AND `posts`.`title` LIKE ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 10 {
		t.Errorf("Expected first param to be int64(10), got %T %v", params[0], params[0])
	}
	if params[1] != "%golang%" {
		t.Errorf("Expected second param to be '%%golang%%', got %v", params[1])
	}
}

func TestComparisonOperators(t *testing.T) {
	// Test various comparison operators
	query := Select(user.ID, user.Name).
		From(user.Table.Name()).
		Where(
			user.ID.Gt(10),
			user.ID.Lt(20),
			user.Name.Like("%John%"),
			user.Email.In("john@example.com", "jane@example.com"),
		)

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "SELECT `users`.`id`, `users`.`name` FROM `users` WHERE `users`.`id` > ? AND `users`.`id` < ? AND `users`.`name` LIKE ? AND `users`.`email` IN (?, ?)"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 5 {
		t.Errorf("Expected 5 params, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 10 {
		t.Errorf("Expected first param to be int64(10), got %T %v", params[0], params[0])
	}
	if v, ok := params[1].(int64); !ok || v != 20 {
		t.Errorf("Expected second param to be int64(20), got %T %v", params[1], params[1])
	}
	if params[2] != "%John%" {
		t.Errorf("Expected third param to be '%%John%%', got %v", params[2])
	}
	if params[3] != "john@example.com" {
		t.Errorf("Expected fourth param to be 'john@example.com', got %v", params[3])
	}
	if params[4] != "jane@example.com" {
		t.Errorf("Expected fifth param to be 'jane@example.com', got %v", params[4])
	}
}

func TestStringOperations(t *testing.T) {
	// Test string operations: contains, startsWith, endsWith
	query := Select(user.ID, user.Name).
		From(user.Table.Name()).
		Where(
			user.Name.Contains("John"),
			user.Email.StartsWith("john"),
			user.Email.EndsWith("example.com"),
		)

	sqlStr, _, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	// Just check if the SQL contains the expected parts
	if !strings.Contains(sqlStr, "LIKE ?") {
		t.Errorf("Expected SQL to contain 'LIKE ?', got: %s", sqlStr)
	}

	// Check limit/offset features
	limitOnlyQuery := Select(user.ID).
		From(user.Table.Name()).
		Limit(10)

	sqlStr, _, err = limitOnlyQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	if !strings.Contains(sqlStr, "LIMIT 10") {
		t.Errorf("Expected SQL to contain 'LIMIT 10', got: %s", sqlStr)
	}

	offsetOnlyQuery := Select(user.ID).
		From(user.Table.Name()).
		Offset(5)

	sqlStr, _, err = offsetOnlyQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	if !strings.Contains(sqlStr, "OFFSET 5") {
		t.Errorf("Expected SQL to contain 'OFFSET 5', got: %s", sqlStr)
	}
}

// Helper function to create a field.OrderField for a column name
func fieldAsc(name string) field.OrderField {
	// Create a StringField with empty table (works for aliases in ORDER BY)
	f := field.StringField{FieldName: name, TableName: ""}
	return f.Asc()
}

func fieldDesc(name string) field.OrderField {
	// Create a StringField with empty table (works for aliases in ORDER BY)
	f := field.StringField{FieldName: name, TableName: ""}
	return f.Desc()
}

func TestAggregatesAndGroupBy(t *testing.T) {
	// Test GROUP BY and aggregate functions
	query := Select(user.ID, Count(post.ID).As("post_count")).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID)).
		GroupBy(user.ID)

	sqlStr, params, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "SELECT `users`.`id`, COUNT(`posts`.`id`) AS `post_count` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` GROUP BY `users`.`id`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 0 {
		t.Errorf("Expected 0 params, got %d", len(params))
	}

	// Test with HAVING clause
	queryWithHaving := Select(user.ID, Count(post.ID).As("post_count")).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID)).
		GroupBy(user.ID).
		Having(Count(post.ID).Gt(5))

	sqlStr, params, err = queryWithHaving.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL = "SELECT `users`.`id`, COUNT(`posts`.`id`) AS `post_count` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` GROUP BY `users`.`id` HAVING COUNT(`posts`.`id`) > ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(params))
	}
	if v, ok := params[0].(int64); !ok || v != 5 {
		t.Errorf("Expected param to be int64(5), got %T %v", params[0], params[0])
	}

	// Test a complex query with all features
	complexQuery := Select(
		user.ID, user.Name,
		post.ID, post.Title,
		Count(comment.ID).As("comment_count"),
	).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID)).
		LeftJoin(comment.Table.Name(), post.ID.EqField(comment.PostID)).
		Where(
			user.ID.Gt(10),
			post.Title.Like("%golang%"),
		).
		GroupBy(user.ID, user.Name, post.ID, post.Title).
		Having(Count(comment.ID).Gt(2)).
		OrderBy(fieldDesc("comment_count")).
		Limit(10).
		Offset(20)

	sqlStr, params, err = complexQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedComplexSQL := "SELECT `users`.`id`, `users`.`name`, `posts`.`id`, `posts`.`title`, COUNT(`comments`.`id`) AS `comment_count` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` LEFT JOIN `comments` ON `posts`.`id` = `comments`.`post_id` WHERE `users`.`id` > ? AND `posts`.`title` LIKE ? GROUP BY `users`.`id`, `users`.`name`, `posts`.`id`, `posts`.`title` HAVING COUNT(`comments`.`id`) > ? ORDER BY ``.`comment_count` DESC LIMIT 20,10"
	if sqlStr != expectedComplexSQL {
		t.Errorf("Expected complex SQL: %s, got: %s", expectedComplexSQL, sqlStr)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}

	if v, ok := params[0].(int64); !ok || v != 10 {
		t.Errorf("Expected first param to be int64(10), got %T %v", params[0], params[0])
	}

	if params[1] != "%golang%" {
		t.Errorf("Expected second param to be '%%golang%%', got %v", params[1])
	}

	if v, ok := params[2].(int64); !ok || v != 2 {
		t.Errorf("Expected third param to be int64(2), got %T %v", params[2], params[2])
	}
}

func TestFieldAliases(t *testing.T) {
	// Test field aliases
	query := Select(
		user.ID.As("user_id"),
		user.Name.As("user_name"),
	).
		From(user.Table.Name())

	sqlStr, _, err := query.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL := "SELECT `users`.`id` AS `user_id`, `users`.`name` AS `user_name` FROM `users`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// Test join with field aliases
	joinQuery := Select(
		user.ID.As("user_id"),
		user.Name.As("user_name"),
		post.Title.As("post_title"),
	).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID))

	sqlStr, _, err = joinQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL = "SELECT `users`.`id` AS `user_id`, `users`.`name` AS `user_name`, `posts`.`title` AS `post_title` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// Test aggregate function with alias
	aggregateQuery := Select(
		user.ID.As("user_id"),
		Count(post.ID).As("post_count"),
	).
		From(user.Table.Name()).
		Join(post.Table.Name(), user.ID.EqField(post.UserID)).
		GroupBy(user.ID)

	sqlStr, _, err = aggregateQuery.SQL()
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	expectedSQL = "SELECT `users`.`id` AS `user_id`, COUNT(`posts`.`id`) AS `post_count` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` GROUP BY `users`.`id`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}
}
