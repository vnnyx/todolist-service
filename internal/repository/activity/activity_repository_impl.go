package activity

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepositoryImpl struct {
	db             *sql.DB
	cache          *cache.Cache
	workerActivity chan *entity.Activity
	memdb          *memdb.MemDB
	mutex          sync.Mutex
}

type cacheKey struct {
	key string
}

func NewActivityRepository() ActivityRepository {
	return &ActivityRepositoryImpl{
		workerActivity: make(chan *entity.Activity),
		mutex:          sync.Mutex{},
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

func (repo *ActivityRepositoryImpl) InjectMemDB(memdb *memdb.MemDB) error {
	repo.memdb = memdb
	return nil
}

func (repo *ActivityRepositoryImpl) Worker(ctx context.Context) {
	for {
		query := "INSERT INTO activities(activity_id, title, email, created_at, updated_at) VALUES(?,?,?,?,?)"
		select {
		case <-ctx.Done():
			return // exit if the context is cancelled
		case activity := <-repo.workerActivity:
			tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
			if err != nil {
				log.Fatalf("error starting transaction: %v", err)
			}

			args := []interface{}{
				activity.ID,
				activity.Title,
				activity.Email,
				activity.CreatedAt,
				activity.UpdatedAt,
			}

			// Use a context with timeout for the insert operation
			insertCtx, insertCancel := context.WithTimeout(ctx, time.Second*10)
			defer insertCancel()

			stmt, err := tx.PrepareContext(ctx, query)
			if err != nil {
				// rollback the transaction if an error occurs
				tx.Rollback()
				return
			}
			defer stmt.Close()

			_, err = stmt.ExecContext(insertCtx, args...)
			if err != nil {
				// rollback the transaction if an error occurs
				tx.Rollback()
				return
			}
			tx.Commit()

			// insert the activity into the memdb
			txn := repo.memdb.Txn(true)
			defer txn.Abort()
			err = txn.Insert("activities", activity)
			if err != nil {
				return
			}
			txn.Commit()
		}
	}
}

func (repo *ActivityRepositoryImpl) InsertActivity(activity *entity.Activity) error {
	go func() {
		repo.cache.Flush()
	}()

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()
	activity.ID = entity.ActivitySeq
	entity.ActivitySeq++

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(activity *entity.Activity) {
		repo.workerActivity <- activity
		wg.Done()
	}(activity)

	wg.Wait()

	return nil
}

func (repo *ActivityRepositoryImpl) GetActivityByID(id int64) (activity *entity.Activity, err error) {
	key := cacheKey{"activityId-" + strconv.FormatInt(id, 10)}
	repo.mutex.Lock()
	data, found := repo.cache.Get(key.key)
	repo.mutex.Unlock()
	if !found {
		ctx, cancel := infrastructure.NewMySQLContext()
		defer cancel()

		query := "SELECT * FROM activities WHERE activity_id=?"
		stmt, err := repo.db.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.QueryContext(ctx, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			var a = new(entity.Activity)
			err := rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
			if err != nil {
				return nil, err
			}
			repo.mutex.Lock()
			repo.cache.SetDefault(key.key, a)
			repo.mutex.Unlock()
			return a, nil
		}
		return nil, fmt.Errorf("Activity with ID %v Not Found", id)
	}
	return data.(*entity.Activity), nil
}

func (repo *ActivityRepositoryImpl) GetAllActivity() (activities []*entity.Activity, err error) {
	// Check the cache first
	data, found := repo.cache.Get("allactivity")
	if found {
		return data.([]*entity.Activity), nil
	}

	// Acquire a mutex to ensure only one goroutine executes the database query
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Check the cache again to make sure that another goroutine hasn't already executed the database query
	data, found = repo.cache.Get("allactivity")
	if found {
		return data.([]*entity.Activity), nil
	}

	// If the data is not found in the cache, execute the database query
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

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

	// Update the cache with the data
	repo.cache.SetDefault("allactivity", activities)

	return activities, nil
}

func (repo *ActivityRepositoryImpl) UpdateActivity(activity *entity.Activity) error {
	go func() {
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE activities SET title=? WHERE activity_id=?"

	// Prepare the SQL statement
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the prepared statement with the given parameters
	_, err = stmt.ExecContext(ctx, activity.Title, activity.ID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ActivityRepositoryImpl) DeleteActivity(id int64) error {
	go func() {
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "DELETE FROM activities WHERE activity_id=?"

	// Prepare the SQL statement
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the prepared statement with the given parameter
	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
