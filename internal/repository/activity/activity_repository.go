package activity

import (
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepository interface {
	InsertActivity(activity entity.Activity) (*entity.Activity, error)
	GetActivityByID(id int64) (activity *entity.Activity, err error)
	GetAllActivity() (activities []*entity.Activity, err error)
	UpdateActivity(activity entity.Activity) (*entity.Activity, error)
	DeleteActivity(id int64) error
}
