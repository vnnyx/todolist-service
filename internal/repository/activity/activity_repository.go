package activity

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-memdb"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepository interface {
	InsertActivity(activity *entity.Activity) error
	GetActivityByID(id int64) (activity *entity.Activity, err error)
	GetAllActivity() (activities []*entity.Activity, err error)
	UpdateActivity(activity *entity.Activity) error
	DeleteActivity(id int64) error
	Worker(ctx context.Context)

	//dependency injection
	InjectDB(db *sql.DB) error
	InjectMemDB(memdb *memdb.MemDB) error
	InjectCache(cache *cache.Cache) error
}
