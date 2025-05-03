package comment

import (
	"time"

	"github.com/xhd2015/arc-orm/example/engine"
	"github.com/xhd2015/arc-orm/orm"
	"github.com/xhd2015/arc-orm/table"
)

// Table is the comments table
var Table = table.New("comments")

var ORM = orm.Bind[Comment, CommentOptional](engine.GetEngine(), Table)

// Field definitions
var (
	ID         = Table.Int64("id")
	Content    = Table.String("content")
	PostID     = Table.Int64("post_id")
	Score      = Table.Float64("score")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

//go:generate go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest gen
type Comment struct {
	Id         int64
	Content    string
	PostId     int64
	Score      float64
	CreateTime time.Time
	UpdateTime time.Time
}

type CommentOptional struct {
	Id         *int64
	Content    *string
	PostId     *int64
	Score      *float64
	CreateTime *time.Time
	UpdateTime *time.Time
}
