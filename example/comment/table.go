package comment

import (
	"time"

	"github.com/xhd2015/ormx/example/engine"
	"github.com/xhd2015/ormx/orm"
	"github.com/xhd2015/ormx/table"
)

// Table is the comments table
var Table = table.New("comments")

var ORM = orm.Bind[Comment, CommentOptional](engine.GetEngine(), Table)

// Field definitions
var (
	ID         = Table.Int64("id")
	Content    = Table.String("content")
	PostID     = Table.Int64("post_id")
	CreateTime = Table.Time("create_time")
	UpdateTime = Table.Time("update_time")
)

type Comment struct {
	Id int64
	Content string
	PostId int64
	CreateTime time.Time
	UpdateTime time.Time
}

type CommentOptional struct {
	Id *int64
	Content *string
	PostId *int64
	CreateTime *time.Time
	UpdateTime *time.Time
}
