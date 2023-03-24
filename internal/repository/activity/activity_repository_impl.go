package activity

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepositoryImpl struct {
	db    *sql.DB
	cache *cache.Cache
	memdb *memdb.MemDB
}

func NewActivityRepository(db *sql.DB, cache *cache.Cache, memdb *memdb.MemDB) ActivityRepository {
	return &ActivityRepositoryImpl{
		db:    db,
		cache: cache,
		memdb: memdb,
	}
}

func (repo *ActivityRepositoryImpl) InsertActivity(activity entity.Activity) (*entity.Activity, error) {
	var wg sync.WaitGroup
	var id = make(chan int64, 1)

	wg.Add(3)
	go func() {
		defer wg.Done()
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	go func(activity entity.Activity) {
		defer wg.Done()
		query := "INSERT INTO activities(title, email, created_at, updated_at) VALUES(?,?,?,?)"

		args := []interface{}{
			activity.Title,
			activity.Email,
			activity.CreatedAt,
			activity.UpdatedAt,
		}
		result, err := repo.db.ExecContext(ctx, query, args...)
		if err != nil {
			logrus.Error(err)
			return
		}
		ids, err := result.LastInsertId()
		if err != nil {
			logrus.Error(err)
			return
		}
		id <- ids
		close(id)
	}(activity)
	activity.ID = <-id
	go func(activity entity.Activity) {
		defer wg.Done()
		txn := repo.memdb.Txn(true)
		defer txn.Abort()
		err := txn.Insert("activities", activity)
		if err != nil {
			logrus.Error(err)
			return
		}
		txn.Commit()
	}(activity)
	wg.Wait()

	return &activity, nil
}

func (repo *ActivityRepositoryImpl) GetActivityByID(id int64) (activity *entity.Activity, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	data, found := repo.cache.Get(fmt.Sprintf("id-%v", id))
	if !found {
		txn := repo.memdb.Txn(false)
		defer txn.Abort()
		raw, err := txn.First("activities", "id", id)
		if err != nil {
			return nil, err
		}
		if raw != nil {
			a := raw.(entity.Activity)
			return &a, nil
		}
		query := "SELECT * FROM activities WHERE activity_id=?"
		rows, err := repo.db.QueryContext(ctx, query, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			var a entity.Activity
			err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
			if err != nil {
				return nil, err
			}
			repo.cache.Set(fmt.Sprintf("id-%v", id), a, cache.DefaultExpiration)
			return &a, nil
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
		txn := repo.memdb.Txn(false)
		defer txn.Abort()
		it, err := txn.Get("activities", "id")
		for obj := it.Next(); obj != nil; obj = it.Next() {
			a := obj.(entity.Activity)
			activities = append(activities, &a)
		}

		if len(activities) > 0 {
			repo.cache.Set("allactivity", activities, cache.DefaultExpiration)
			return activities, nil
		}

		query := "SELECT * FROM activities"
		rows, err := repo.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var a entity.Activity
			err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
			if err != nil {
				return nil, err
			}
			activities = append(activities, &a)
		}
		repo.cache.Set("allactivity", activities, cache.DefaultExpiration)
		return activities, nil
	}

	return data.([]*entity.Activity), nil
}

func (repo *ActivityRepositoryImpl) UpdateActivity(activity entity.Activity) (*entity.Activity, error) {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	activity.UpdatedAt = time.Now()

	query := "UPDATE activities SET title=?, updated_at=? WHERE activity_id=?"
	args := []interface{}{
		activity.Title,
		activity.UpdatedAt,
		activity.ID,
	}
	_, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &activity, nil
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
