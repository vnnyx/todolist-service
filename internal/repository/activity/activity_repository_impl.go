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
	mutex          sync.RWMutex
}

type cacheKey struct {
	key string
}

func NewActivityRepository() ActivityRepository {
	return &ActivityRepositoryImpl{
		workerActivity: make(chan *entity.Activity),
		mutex:          sync.RWMutex{},
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

			repo.cache.Flush()
		}
	}
}

func (repo *ActivityRepositoryImpl) InsertActivity(activity *entity.Activity) error {
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

func (repo *ActivityRepositoryImpl) GetActivityByID(id int64) (*entity.Activity, error) {
	// Create a cache key for this activity ID
	key := "activityId-" + strconv.FormatInt(id, 10)

	// Try to retrieve the activity from the cache
	repo.mutex.Lock()
	data, found := repo.cache.Get(key)
	repo.mutex.Unlock()
	if found {
		return data.(*entity.Activity), nil
	}

	// Try to retrieve the activity from the memdb
	if activity, err := repo.getActivityByIDFromMemDB(id); err == nil {
		// If the activity was found in memdb, add it to the cache and return it
		repo.cache.SetDefault(key, activity)
		return activity, nil
	}

	// If the activity wasn't found in the memdb, retrieve it from the database
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "SELECT * FROM activities WHERE activity_id=?"
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var activity entity.Activity
	err = row.Scan(&activity.ID, &activity.Title, &activity.Email, &activity.CreatedAt, &activity.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Activity with ID %v Not Found", id)
	}
	if err != nil {
		return nil, err
	}

	// Add the activity to the cache and return it
	repo.mutex.Lock()
	repo.cache.SetDefault(key, &activity)
	repo.mutex.Unlock()
	return &activity, nil
}

func (repo *ActivityRepositoryImpl) GetAllActivity() ([]*entity.Activity, error) {
	// Check cache first
	key := "allactivity"
	data, found := repo.cache.Get(key)
	if found {
		return data.([]*entity.Activity), nil
	}

	// Acquire a read lock to query memdb
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// Check the cache again to make sure that another goroutine hasn't already executed the database query
	data, found = repo.cache.Get(key)
	if found {
		return data.([]*entity.Activity), nil
	}

	// Query memdb concurrently
	activitiesCh := make(chan *entity.Activity, 10)
	go func() {
		defer close(activitiesCh)
		txn := repo.memdb.Txn(false)
		defer txn.Abort()
		it, _ := txn.Get("activities", "id")
		for obj := it.Next(); obj != nil; obj = it.Next() {
			activity := obj.(*entity.Activity)
			activitiesCh <- activity
		}
	}()

	// Collect memdb results and check if found
	var activities []*entity.Activity
	for activity := range activitiesCh {
		activities = append(activities, activity)
	}
	if len(activities) > 0 {
		repo.cache.SetDefault(key, activities)
		return activities, nil
	}

	// Query MySQL and update cache
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()
	rows, err := repo.db.QueryContext(ctx, "SELECT * FROM activities")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var activity entity.Activity
		err := rows.Scan(&activity.ID, &activity.Title, &activity.Email, &activity.CreatedAt, &activity.UpdatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}

	repo.cache.SetDefault(key, activities)

	return activities, nil
}

func (repo *ActivityRepositoryImpl) UpdateActivity(activity *entity.Activity) error {
	go func() {
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE activities SET title=?, updated_at=? WHERE activity_id=?"

	// Prepare the SQL statement
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	activity.UpdatedAt = time.Now()

	args := []interface{}{
		activity.Title,
		activity.UpdatedAt,
		activity.ID,
	}

	// Execute the prepared statement with the given parameters
	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	// Update the memdb
	txn := repo.memdb.Txn(true)
	defer txn.Abort()

	// Get the activity from memdb
	oldActivity, err := repo.getActivityByIDFromMemDB(activity.ID)
	if err != nil {
		return err
	}

	// Update the fields that have changed
	if oldActivity.Title != activity.Title {
		oldActivity.Title = activity.Title
	}
	if oldActivity.Email != activity.Email {
		oldActivity.Email = activity.Email
	}
	if oldActivity.UpdatedAt != activity.UpdatedAt {
		oldActivity.UpdatedAt = activity.UpdatedAt
	}

	// Update the todo in memdb
	err = txn.Insert("activities", oldActivity)
	if err != nil {
		return err
	}

	txn.Commit()

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

	// Delete the activity in memdb
	txn := repo.memdb.Txn(true)
	defer txn.Abort()
	err = txn.Delete("activities", entity.Activity{ID: id})
	if err != nil {
		return err
	}
	txn.Commit()

	return nil
}

func (repo *ActivityRepositoryImpl) getActivityByIDFromMemDB(activityID int64) (*entity.Activity, error) {
	txn := repo.memdb.Txn(false)
	defer txn.Abort()

	got, err := txn.First("activities", "id", activityID)
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, fmt.Errorf("activity with ID %d not found in memdb", activityID)
	}

	return got.(*entity.Activity), nil
}
