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
