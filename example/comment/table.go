package comment

import "github.com/xhd2015/ormx/table"

// Table is the comments table
var Table = table.New("comments")

// Field definitions
var (
	ID        = Table.Int64("id")
	Content   = Table.String("content")
	PostID    = Table.Int64("post_id")
	CreatedAt = Table.Time("created_at")
)
