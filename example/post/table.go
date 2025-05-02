package post

import (
	"time"

	"github.com/xhd2015/ormx/example/engine"
	"github.com/xhd2015/ormx/orm"
	"github.com/xhd2015/ormx/table"
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

var ORM = orm.Bind[Post, PostOptional](engine.GetEngine(), Table)

type Post struct {
	Id int64
	Title string
	UserId int64
	CreateTime time.Time
	UpdateTime time.Time
}

type PostOptional struct {
	Id *int64
	Title *string
	UserId *int64
	CreateTime *time.Time
	UpdateTime *time.Time
}
