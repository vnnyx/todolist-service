package todo

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

type TodoRepositoryImpl struct {
	db         *sql.DB
	cache      *cache.Cache
	workerTodo chan *entity.Todo
	memdb      *memdb.MemDB
	mutex      sync.RWMutex
}

type cacheKey struct {
	key string
}

func NewTodoRepository() TodoRepository {
	return &TodoRepositoryImpl{
		workerTodo: make(chan *entity.Todo),
		mutex:      sync.RWMutex{},
	}
}

func (repo *TodoRepositoryImpl) InjectDB(db *sql.DB) error {
	repo.db = db
	return nil
}

func (repo *TodoRepositoryImpl) InjectCache(cache *cache.Cache) error {
	repo.cache = cache
	return nil
}

func (repo *TodoRepositoryImpl) InjectMemDB(memdb *memdb.MemDB) error {
	repo.memdb = memdb
	return nil
}

func (repo *TodoRepositoryImpl) Worker(ctx context.Context) {
	for {
		query := "INSERT INTO todos(todo_id, activity_group_id, title, is_active, priority, created_at, updated_at) VALUES(?,?,?,?,?,?,?)"
		select {
		case <-ctx.Done():
			return // exit if the context is cancelled
		case todo := <-repo.workerTodo:
			tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
			if err != nil {
				log.Fatalf("error starting transaction: %v", err)
			}

			args := []interface{}{
				todo.ID,
				todo.ActivityGroupID,
				todo.Title,
				todo.IsActive,
				todo.Priority,
				todo.CreatedAt,
				todo.UpdatedAt,
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

			// insert the todo into the memdb
			txn := repo.memdb.Txn(true)
			defer txn.Abort()
			err = txn.Insert("todos", todo)
			if err != nil {
				return
			}
			txn.Commit()

			// clear the cache after inserting
			repo.cache.Flush()
		}
	}
}

func (repo *TodoRepositoryImpl) InsertTodo(todo *entity.Todo) error {
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	todo.ID = entity.TodoSeq
	entity.TodoSeq++

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(todo *entity.Todo) {
		repo.workerTodo <- todo
		wg.Done()
	}(todo)

	wg.Wait()

	return nil
}

func (repo *TodoRepositoryImpl) GetTodoByID(id int64) (*entity.Todo, error) {
	// Create a cache key for this todo ID
	key := "todoId-" + strconv.FormatInt(id, 10)

	// Try to retrieve the todo from the cache
	repo.mutex.Lock()
	data, found := repo.cache.Get(key)
	repo.mutex.Unlock()
	if found {
		return data.(*entity.Todo), nil
	}

	// Try to retrieve the todo from the memdb
	if todo, err := repo.getTodoByIDFromMemDB(id); err == nil {
		// If the todo was found in memdb, add it to the cache and return it
		repo.mutex.Lock()
		repo.cache.SetDefault(key, todo)
		repo.mutex.Unlock()
		return todo, nil
	}

	// If the todo wasn't found in the memdb, retrieve it from the database
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "SELECT * FROM todos WHERE todo_id=?"
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var todo entity.Todo
	err = row.Scan(&todo.ID, &todo.ActivityGroupID, &todo.Title, &todo.IsActive, &todo.Priority, &todo.CreatedAt, &todo.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Todo with ID %v Not Found", id)
	}
	if err != nil {
		return nil, err
	}

	// Add the todo to the cache and return it
	repo.mutex.Lock()
	repo.cache.SetDefault(key, &todo)
	repo.mutex.Unlock()
	return &todo, nil
}

func (repo *TodoRepositoryImpl) GetAllTodo(activityGroupID int64) ([]*entity.Todo, error) {
	// Check cache first
	key := cacheKey{"alltodo-" + strconv.FormatInt(activityGroupID, 10)}
	repo.mutex.Lock()
	data, found := repo.cache.Get(key.key)
	repo.mutex.Unlock()
	if found {
		return data.([]*entity.Todo), nil
	}

	// If not found in cache, query memdb
	todos := make([]*entity.Todo, 0)
	txn := repo.memdb.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	if activityGroupID == 0 {
		it, _ = txn.Get("todos", "id")
	} else {
		it, _ = txn.Get("todos", "activity_group_id", activityGroupID)
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		todo := obj.(*entity.Todo)
		todos = append(todos, todo)
	}

	// If found in memdb, update cache and return
	if len(todos) > 0 {
		repo.mutex.Lock()
		repo.cache.SetDefault(key.key, todos)
		repo.mutex.Unlock()
		return todos, nil
	}

	// If not found in memdb, query MySQL
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()
	rows, err := repo.db.QueryContext(ctx, "SELECT * FROM todos WHERE activity_group_id = ?", activityGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		todo := new(entity.Todo)
		err := rows.Scan(&todo.ID, &todo.ActivityGroupID, &todo.Title, &todo.IsActive, &todo.Priority, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	// Update memdb and cache with MySQL results
	txn = repo.memdb.Txn(true)
	defer txn.Abort()
	for _, todo := range todos {
		err := txn.Insert("todos", todo)
		if err != nil {
			return nil, err
		}
	}
	repo.mutex.Lock()
	repo.cache.SetDefault(key.key, todos)
	repo.mutex.Unlock()

	return todos, nil
}

func (repo *TodoRepositoryImpl) UpdateTodo(todo *entity.Todo) error {
	go func() {
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE todos SET title=?, priority=?, is_active=?, updated_at=? WHERE todo_id=?"

	// Prepare the SQL statement
	stmt, err := repo.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	todo.UpdatedAt = time.Now()

	args := []interface{}{
		todo.Title,
		todo.Priority,
		todo.IsActive,
		todo.UpdatedAt,
		todo.ID,
	}

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	// Update the memdb
	txn := repo.memdb.Txn(true)
	defer txn.Abort()

	// Get the todo from memdb
	oldTodo, err := repo.getTodoByIDFromMemDB(todo.ID)
	if err != nil {
		return err
	}

	// Update the fields that have changed
	if oldTodo.Title != todo.Title {
		oldTodo.Title = todo.Title
	}
	if oldTodo.Priority != todo.Priority {
		oldTodo.Priority = todo.Priority
	}
	if oldTodo.IsActive != todo.IsActive {
		oldTodo.IsActive = todo.IsActive
	}
	if oldTodo.UpdatedAt != todo.UpdatedAt {
		oldTodo.UpdatedAt = todo.UpdatedAt
	}

	// Update the todo in memdb
	err = txn.Insert("todos", oldTodo)
	if err != nil {
		return err
	}

	txn.Commit()

	return nil
}

func (repo *TodoRepositoryImpl) DeleteTodo(id int64, title string) error {
	// Flush the cache asynchronously
	go func() {
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	// Begin a transaction
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare the delete statement
	stmt, err := tx.PrepareContext(ctx, "DELETE FROM todos WHERE todo_id=? AND title=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the delete statement
	result, err := stmt.ExecContext(ctx, id, title)
	if err != nil {
		return err
	}

	// Check the number of affected rows
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("Todo with ID %v and Title: %v Not Found", id, title)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// Delete the todo in memdb
	txn := repo.memdb.Txn(true)
	defer txn.Abort()
	err = txn.Delete("todos", entity.Todo{ID: id})
	if err != nil {
		return err
	}
	txn.Commit()

	return nil
}

func (repo *TodoRepositoryImpl) getTodoByIDFromMemDB(todoID int64) (*entity.Todo, error) {
	txn := repo.memdb.Txn(false)
	defer txn.Abort()

	got, err := txn.First("todos", "id", todoID)
	if err != nil {
		return nil, err
	}
	if got == nil {
		return nil, fmt.Errorf("todo with ID %d not found in memdb", todoID)
	}

	return got.(*entity.Todo), nil
}
