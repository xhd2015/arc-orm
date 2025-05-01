# ormx

A type-safe SQL query builder for Go with compile-time type checking.

## Features

- Compile-time type safety for database operations
- Fluent API for building SQL queries
- Support for SELECT, UPDATE, JOIN operations
- Type-safe field expressions and literals
- Clean, concise syntax

## Installation

```bash
go get github.com/xhd2015/ormx
```

## Usage

### Table and Columns Definitions

Define tables and columns in separate packages for better organization:

```go
// Package user defines user table schema
package user

import "github.com/xhd2015/ormx/table"

// Table is the users table
var Table = table.New("users")

// Field definitions
var (
    ID   = Table.Int64("id")
    Name = Table.String("name")
    Age  = Table.Int64("age")
    Email = Table.String("email")
)
```

```go
// Package post defines post table schema
package post

import "github.com/xhd2015/ormx/table"

// Table is the posts table
var Table = table.New("posts")

// Field definitions
var (
    ID        = Table.Int64("id")
    Title     = Table.String("title")
    UserID    = Table.Int64("user_id")
    CreatedAt = Table.Time("created_at")
)
```

### Building Queries

```go
// Import the table definition packages
import (
    "github.com/xhd2015/ormx/sql"
    "github.com/example/myapp/user"
    "github.com/example/myapp/post"
)

// SELECT query with JOIN
query, args, err := sql.
    Select(user.ID, user.Name, post.Title).
    From(user.Table).
    Join(post.Table, user.ID.EqField(post.UserID)).
    Where(user.Age.Gt(18)).
    OrderBy(post.CreatedAt.Desc()).
    Limit(10).
    SQL()
// Output: 
// query = "SELECT `users`.`id`, `users`.`name`, `posts`.`title` FROM `users` JOIN `posts` ON `users`.`id` = `posts`.`user_id` WHERE `users`.`age` > ? ORDER BY `posts`.`created_at` DESC LIMIT ?"
// args = [18, 10]

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
```


## License

MIT 