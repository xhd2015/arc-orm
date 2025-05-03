# arc-orm

A feature rich type-safe SQL query builder for Go with compile-time type checking.

With a simple ORM binding model generator.

## Features

- Compile-time type safety for database operations
- Fluent API for building SQL queries
- Support for SELECT, UPDATE, DELETE, INSERT operations
- Type-safe field expressions and literals
- Clean, concise syntax

## Installation

Library:
```sh
go get github.com/xhd2015/arc-orm
```

ORM model generator `arc-orm gen`:
```sh
# update ORM models according to defined fields
go install github.com/xhd2015/arc-orm/cmd/arc-orm@latest
arc-orm gen
```

## Usage

### Table and Columns Definitions

Define tables and columns in separate packages for better organization:

```go
// Package user defines user table schema
package user

import (
    "time"
    
    "github.com/xhd2015/arc-orm/example/engine"
    "github.com/xhd2015/arc-orm/orm"
    "github.com/xhd2015/arc-orm/table"
)

// Table is the users table
var Table = table.New("users")

// Field definitions
var (
    ID         = Table.Int64("id")
    Name       = Table.String("name")
    Email      = Table.String("email")
    Age        = Table.Int64("age")
    CreateTime = Table.Time("create_time")
    UpdateTime = Table.Time("update_time")
)

// bind ORM with model and table
var ORM = orm.Bind[User, UserOptional](engine.Engine, Table)

// the following two models can be automatically generated and updated

//go:generate go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest gen
// User model that matches the table structure
type User struct {
    Id         int64
    Name       string
    Age        int64
    Email      string
    CreateTime time.Time
    UpdateTime time.Time
}

// UserOptional for partial updates (all fields are pointers)
type UserOptional struct {
    Id         *int64
    Name       *string
    Age        *int64
    Email      *string
    CreateTime *time.Time
    UpdateTime *time.Time
}
```

```go
// Package post defines post table schema
package post

import (
    "time"
    
    "github.com/xhd2015/arc-orm/example/engine"
    "github.com/xhd2015/arc-orm/orm"
    "github.com/xhd2015/arc-orm/table"
)

// Table is the posts table
var Table = table.New("posts")

// Field definitions
var (
    ID         = Table.Int64("id")
    Title      = Table.String("title")
    UserID     = Table.Int64("user_id")
    CreateTime = Table.Time("create_time")
    UpdateTime = Table.Time("update_time")
)

// bind ORM with model and table
var ORM = orm.Bind[Post, PostOptional](engine.Engine, Table)

// the following two models can be automatically generated and updated

//go:generate go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest gen
// Post model that matches the table structure
type Post struct {
    Id         int64
    Title      string
    UserId     int64
    CreateTime time.Time
    UpdateTime time.Time
}

// PostOptional for partial updates (all fields are pointers)
type PostOptional struct {
    Id         *int64
    Title      *string
    UserId     *int64
    CreateTime *time.Time
    UpdateTime *time.Time
}
```

### Define Engine Adaptor
```go
package engine

import "github.com/xhd2015/arc-orm/engine"

var Engine = engine.Getter(get)

var engineImpl engine.Engine

func Init(impl engine.Engine) {
	engineImpl = impl
}

func get() engine.Engine {
	return engineImpl
}
```

### Query with ORM

Once you've defined your table structure and created an ORM instance, you can use it to perform various database operations:

```go
import (
    "context"
    "log"
    "time"
    
    "github.com/xhd2015/arc-orm/sql"
    "github.com/example/myapp/user"
    "github.com/example/myapp/post"
)

func main() {
    ctx := context.Background()
    
    // Query users by ID
    userRecord, err := user.ORM.GetByID(ctx, 123)
    if err != nil {
        log.Fatalf("Failed to query user: %v", err)
    }
    if userRecord == nil {
        log.Println("User not found")
    } else {
        log.Printf("Found user: %s (ID: %d)", userRecord.Name, userRecord.ID)
    }
    
    // Insert a new user
    newUser := &user.User{
        Name:  "Jane Smith",
        Email: "jane@example.com",
        Age:   28,
        // CreateTime and UpdateTime will be automatically set to current time
    }
    
    userID, err := user.ORM.Insert(ctx, newUser)
    if err != nil {
        log.Fatalf("Failed to insert user: %v", err)
    }
    log.Printf("Inserted user with ID: %d", userID)
    
    // Update a user partially (only specified fields)
    updateData := &user.UserOptional{
        Name: sql.Ptr("Jane Doe"),
        Email: sql.Ptr("jane.doe@example.com"),
        // Other fields are nil and won't be updated
        // UpdateTime will be automatically updated
    }
    
    err = user.ORM.UpdateByID(ctx, userID, updateData)
    if err != nil {
        log.Fatalf("Failed to update user: %v", err)
    }
    log.Println("User updated successfully")
    
    // Delete a user
    err = user.ORM.DeleteByID(ctx, userID)
    if err != nil {
        log.Fatalf("Failed to delete user: %v", err)
    }
    log.Println("User deleted successfully")
    
    // Insert a post linked to a user
    newPost := &post.Post{
        Title:  "My First Post",
        UserID: 123, // Link to an existing user
        // CreateTime and UpdateTime will be automatically set
    }
    
    postID, err := post.ORM.Insert(ctx, newPost)
    if err != nil {
        log.Fatalf("Failed to insert post: %v", err)
    }
    log.Printf("Inserted post with ID: %d", postID)
    
    // Execute a custom query to find posts by a specific user
    query := "SELECT * FROM posts WHERE user_id = ?"
    args := []interface{}{123}
    
    userPosts, err := post.ORM.Query(ctx, query, args)
    if err != nil {
        log.Fatalf("Failed to query posts: %v", err)
    }
    
    log.Printf("Found %d posts by user", len(userPosts))
    for _, p := range userPosts {
        log.Printf("- Post ID: %d, Title: %s", p.ID, p.Title)
    }
    
    // Execute a count query
    countQuery := "SELECT COUNT(*) as count FROM posts WHERE user_id = ?"
    countArgs := []interface{}{123}
    
    postsWithCount, err := post.ORM.Count(ctx, countQuery, countArgs)
    if err != nil {
        log.Fatalf("Failed to count posts: %v", err)
    }
    
    if len(postsWithCount) > 0 {
        log.Printf("User has %d posts", postsWithCount[0].Count)
    }
}
```

### Using SQL Builders with ORM

You can also combine the SQL builder with ORM operations for more complex queries:

```go
import (
    "context"
    "log"
    
    "github.com/example/myapp/user"
    "github.com/example/myapp/post"
    "github.com/xhd2015/arc-orm/sql"
)

func queryActiveUserPosts(ctx context.Context, userID int64, limit int) ([]*post.Post, error) {
    // Build a complex query using the SQL builder
    query, args, err := sql.
        Select(post.ID, post.Title, post.CreateTime).
        From(post.Table.Name()).
        Join(user.Table.Name(), post.UserID.EqField(user.ID)).
        Where(
            sql.And(
                post.UserID.Eq(userID),
                user.Age.Gt(18),
            ),
        ).
        OrderBy(post.CreateTime.Desc()).
        Limit(limit).
        SQL()
        
    if err != nil {
        return nil, err
    }
    
    // Execute the query using the ORM
    return post.ORM.Query(ctx, query, args)
}
```

This approach allows you to leverage both the type safety of the SQL builder and the convenience of the ORM for database operations.

### Building Raw SQL

```go
// Import the table definition packages
import (
    "github.com/xhd2015/arc-orm/sql"
    "github.com/example/myapp/user"
    "github.com/example/myapp/post"
)

// SELECT query with JOIN
query, args, err := sql.
    Select(user.ID, user.Name, post.Title).
    From(user.Table.Name()).
    Join(post.Table.Name(), user.ID.EqField(post.UserID)).
    Where(user.Age.Gt(18)).
    OrderBy(post.CreatedAt.Desc()).
    Limit(10).
    SQL()
// Output: 
// query = "SELECT `users`.`id`, `users`.`name`, `posts`.`title` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` WHERE `users`.`age` > ? ORDER BY `posts`.`created_at` DESC LIMIT 10"
// args = [18]

// INSERT
insertQuery, insertArgs, err := sql.
    InsertInto(user.Table.Name()).
    Set(user.Name, sql.String("John Doe")).
    Set(user.Email, sql.String("john@example.com")).
    Set(user.Age, sql.Int64(30)).
    SQL()
// Output:
// insertQuery = "INSERT INTO `users` SET `name`=?, `email`=?, `age`=?"
// insertArgs = ["John Doe", "john@example.com", 30]

// UPDATE query with expressions
updateQuery, updateArgs, err := sql.
    Update(user.Table.Name()).
    Set(user.Name, sql.String("John Doe")).
    Set(user.Age, user.Age.Increment(1)).
    Where(user.ID.Eq(1)).
    SQL()
// Output:
// updateQuery = "UPDATE `users` SET `name`=?, `age`=`users`.`age`+? WHERE `users`.`id` = ?"
// updateArgs = ["John Doe", 1, 1]

// DELETE
deleteQuery, deleteArgs, err := sql.
    DeleteFrom(user.Table.Name()).
    Where(user.ID.Eq(1)).
    SQL()
// Output:
// deleteQuery = "DELETE FROM `users` WHERE `users`.`id` = ?"
// deleteArgs = [1]
```

## Integrate with ORMs

This library focuses on building type-safe SQL queries, but doesn't handle query execution or result mapping. Here's how to integrate it with popular Go database query libraries:

### Standard database/sql

Standard database/sql engine adaptor:
```go
package sql_adaptor

import (
	"context"
	"database/sql"
)

type SQLDBEngine struct {
	DB *sql.DB
}

func NewEngine(db *sql.DB) engine.Engine {
	return &SQLDBEngine{DB: db}
}

func (e *SQLDBEngine) Query(ctx context.Context, sqlQuery string, args []interface{}, result interface{}) error {
	// Execute the query
	rows, err := e.DB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Use reflection to populate the result slice
	return scanRowsIntoSlice(rows, result)
}

func (e *SQLDBEngine) Exec(ctx context.Context, sqlQuery string, args []interface{}) error {
	_, err := e.DB.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (e *SQLDBEngine) ExecInsert(ctx context.Context, sqlQuery string, args []interface{}) (int64, error) {
	result, err := e.DB.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Helper functions for reflection operations (scanRowsIntoSlice, makeSlicePtr, hasResults, copyFirstResult)
// would be implemented here
```

### xorm

xorm engine adaptor:
```go
package xorm_adaptor

import (
	"context"

	"xorm.io/xorm"
)

type XormEngine struct {
	Engine *xorm.Engine
}

func NewEngine(engine *xorm.Engine) engine.Engine{
    return &XormEngine{Engine: engine}
}

func (c *XormEngine) Session() *xorm.Session {
	return c.Engine.NoAutoCondition().NoAutoTime()
}

func (e *XormEngine) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	return e.Session().Context(ctx).SQL(sql, args...).Find(result)
}

func (e *XormEngine) Exec(ctx context.Context, sql string, args []interface{}) error {
	_, err := e.Session().Context(ctx).SQL(sql, args...).Exec()
	return err
}

func (e *XormEngine) ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error) {
	res, err := e.Session().Context(ctx).SQL(sql, args...).Exec()
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
```

### sqlx
sqlx engine adaptor:
```go
package sqlx_adaptor

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SQLXEngine struct {
	DB *sqlx.DB
}

func NewEngine(db *sqlx.DB) engine.Engine {
	return &SQLXEngine{DB: db}
}

func (e *SQLXEngine) Query(ctx context.Context, sqlQuery string, args []interface{}, result interface{}) error {
	// Use sqlx's Select method to execute query and populate results
	return e.DB.SelectContext(ctx, result, sqlQuery, args...)
}

func (e *SQLXEngine) Exec(ctx context.Context, sqlQuery string, args []interface{}) error {
	_, err := e.DB.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (e *SQLXEngine) ExecInsert(ctx context.Context, sqlQuery string, args []interface{}) (int64, error) {
	result, err := e.DB.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
```

## License

MIT 

