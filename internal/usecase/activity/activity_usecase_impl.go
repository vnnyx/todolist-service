package activity

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/vnnyx/golang-todo-api/internal/model"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/repository/activity"
)

type ActivityUCImpl struct {
	activityRepository activity.ActivityRepository
}

func NewActivityUC() ActivityUC {
	return &ActivityUCImpl{}
}

func (uc *ActivityUCImpl) InjectActivityRepository(activityRepository activity.ActivityRepository) error {
	uc.activityRepository = activityRepository
	return nil
}

func (uc *ActivityUCImpl) CreateActivity(ctx context.Context, req web.ActivityCreateRequest) (*web.ActivityDTO, error) {
	if req.Title == "" {
		return nil, model.ErrTitleCannotBeNull
	}
	got, err := uc.activityRepository.InsertActivity(entity.Activity{
		Title: req.Title,
		Email: req.Email,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return got.ToDTO(), nil
}

func (uc *ActivityUCImpl) GetActivityByID(ctx context.Context, id int64) (*web.ActivityDTO, error) {
	got, err := uc.activityRepository.GetActivityByID(id)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return got.ToDTO(), nil
}

func (uc *ActivityUCImpl) GetAllActivity(ctx context.Context) ([]*web.ActivityDTO, error) {
	got, err := uc.activityRepository.GetAllActivity()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	res := make([]*web.ActivityDTO, 0)
	for _, a := range got {
		res = append(res, a.ToDTO())
	}

	return res, nil
}

func (uc *ActivityUCImpl) UpdateActivity(ctx context.Context, req web.ActivityUpdateRequest) (*web.ActivityDTO, error) {
	activity, err := uc.activityRepository.GetActivityByID(req.ID)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	activity.Title = req.Title

	got, err := uc.activityRepository.UpdateActivity(*activity)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return got.ToDTO(), nil
}

func (uc *ActivityUCImpl) DeleteActivity(ctx context.Context, id int64) error {
	activity, err := uc.activityRepository.GetActivityByID(id)
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = uc.activityRepository.DeleteActivity(activity.ID)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}
