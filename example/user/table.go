package user

import "github.com/xhd2015/ormx/table"

// Table is the users table
var Table = table.New("users")

// Field definitions
var (
	ID        = Table.Int64("id")
	Name      = Table.String("name")
	Email     = Table.String("email")
	CreatedAt = Table.Time("created_at")
	Age       = Table.Int64("age")
)
