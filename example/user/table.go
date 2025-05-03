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

var ORM = orm.Bind[User, UserOptional](engine.GetEngine(), Table)

//go:generate go run github.com/xhd2015/arc-orm/cmd/arc-orm@latest sync
type User struct {
	Id         int64
	Name       string
	Email      string
	Age        int64
	CreateTime time.Time
	UpdateTime time.Time
}

type UserOptional struct {
	Id         *int64
	Name       *string
	Email      *string
	Age        *int64
	CreateTime *time.Time
	UpdateTime *time.Time
}
