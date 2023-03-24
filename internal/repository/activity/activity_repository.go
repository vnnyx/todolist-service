package activity

import (
	"database/sql"

	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepository interface {
	InsertActivity(activity entity.Activity) (*entity.Activity, error)
	GetActivityByID(id int64) (activity *entity.Activity, err error)
	GetAllActivity() (activities []*entity.Activity, err error)
	UpdateActivity(activity entity.Activity) (*entity.Activity, error)
	DeleteActivity(id int64) error
	Worker()

	//dependency injection
	InjectDB(db *sql.DB) error
	InjectCache(cache *cache.Cache) error
}
