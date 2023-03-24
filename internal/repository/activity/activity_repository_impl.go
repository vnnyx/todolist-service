package activity

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepositoryImpl struct {
	db             *sql.DB
	cache          *cache.Cache
	workerActivity chan entity.Activity
}

func NewActivityRepository() ActivityRepository {
	return &ActivityRepositoryImpl{
		workerActivity: make(chan entity.Activity),
	}
}

func (repo *ActivityRepositoryImpl) InjectDB(db *sql.DB) error {
	repo.db = db
	return nil
}

func (repo *ActivityRepositoryImpl) InjectCache(cache *cache.Cache) error {
	repo.cache = cache
	return nil
}

func (repo *ActivityRepositoryImpl) Worker() {
	for {
		activity := <-repo.workerActivity
		query := "INSERT INTO activities(activity_id, title, email, created_at, updated_at) VALUES(?,?,?,?,?)"

		args := []interface{}{
			activity.ID,
			activity.Title,
			activity.Email,
			activity.CreatedAt,
			activity.UpdatedAt,
		}
		_, err := repo.db.ExecContext(context.Background(), query, args...)
		if err != nil {
			return
		}
	}
}

func (repo *ActivityRepositoryImpl) InsertActivity(activity entity.Activity) (*entity.Activity, error) {
	repo.cache.Flush()

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()
	activity.ID = entity.ActivitySeq
	entity.ActivitySeq++

	go func(activity entity.Activity) {
		repo.workerActivity <- activity
	}(activity)

	return &activity, nil
}

func (repo *ActivityRepositoryImpl) GetActivityByID(id int64) (activity *entity.Activity, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	data, found := repo.cache.Get(fmt.Sprintf("activityId-%v", id))
	if !found {
		query := "SELECT * FROM activities WHERE activity_id=?"
		rows, err := repo.db.QueryContext(ctx, query, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			var a = new(entity.Activity)
			err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
			if err != nil {
				return nil, err
			}
			repo.cache.SetDefault(fmt.Sprintf("activityId-%v", id), a)
			return a, nil
		}
		return nil, fmt.Errorf("Activity with ID %v Not Found", id)
	}

	return data.(*entity.Activity), nil
}

func (repo *ActivityRepositoryImpl) GetAllActivity() (activities []*entity.Activity, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	data, found := repo.cache.Get("allactivity")
	if !found {
		query := "SELECT * FROM activities"
		rows, err := repo.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var a = new(entity.Activity)
			err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
			if err != nil {
				return nil, err
			}
			activities = append(activities, a)
		}
		repo.cache.SetDefault("allactivity", activities)
		return activities, nil
	}
	return data.([]*entity.Activity), nil
}

func (repo *ActivityRepositoryImpl) UpdateActivity(activity entity.Activity) (*entity.Activity, error) {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE activities SET title=? WHERE activity_id=?"
	args := []interface{}{
		activity.Title,
		activity.ID,
	}
	_, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	a, err := repo.GetActivityByID(activity.ID)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (repo *ActivityRepositoryImpl) DeleteActivity(id int64) error {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "DELETE FROM activities WHERE activity_id=?"
	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
