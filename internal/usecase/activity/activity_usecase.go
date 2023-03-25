package activity

import (
	"context"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/repository/activity"
)

type ActivityUC interface {
	CreateActivity(ctx context.Context, req *web.ActivityCreateRequest) (*web.ActivityDTO, error)
	GetActivityByID(ctx context.Context, id int64) (*web.ActivityDTO, error)
	GetAllActivity(ctx context.Context) ([]*web.ActivityDTO, error)
	UpdateActivity(ctx context.Context, req *web.ActivityUpdateRequest) (*web.ActivityDTO, error)
	DeleteActivity(ctx context.Context, id int64) error

	//dependency injection
	InjectActivityRepository(activityRepository activity.ActivityRepository) error
}
