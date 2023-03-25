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
	mutex      sync.Mutex
}

type cacheKey struct {
	key string
}

func NewTodoRepository() TodoRepository {
	return &TodoRepositoryImpl{
		workerTodo: make(chan *entity.Todo),
		mutex:      sync.Mutex{},
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
		}
	}
}

func (repo *TodoRepositoryImpl) InsertTodo(todo *entity.Todo) error {
	go func() {
		repo.cache.Flush()
	}()

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

func (repo *TodoRepositoryImpl) GetTodoByID(id int64) (todo *entity.Todo, err error) {
	key := cacheKey{"todoId-" + strconv.FormatInt(id, 10)}
	repo.mutex.Lock()
	data, found := repo.cache.Get(key.key)
	repo.mutex.Unlock()
	if !found {
		ctx, cancel := infrastructure.NewMySQLContext()
		defer cancel()

		query := "SELECT * FROM todos WHERE todo_id=?"
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
			var t = new(entity.Todo)
			err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
			if err != nil {
				return nil, err
			}
			repo.mutex.Lock()
			repo.cache.SetDefault(key.key, t)
			repo.mutex.Unlock()
			return t, nil
		}
		return nil, fmt.Errorf("Todo with ID %v Not Found", id)
	}
	return data.(*entity.Todo), nil
}

func (repo *TodoRepositoryImpl) GetAllTodo(activityGroupID int64) ([]*entity.Todo, error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	key := cacheKey{"alltodo-" + strconv.FormatInt(activityGroupID, 10)}

	repo.mutex.Lock()
	data, found := repo.cache.Get(key.key)
	repo.mutex.Unlock()

	if found {
		return data.([]*entity.Todo), nil
	}

	query := "SELECT * FROM todos"
	args := make([]interface{}, 0)
	if activityGroupID != 0 {
		query += " WHERE activity_group_id = ?"
		args = append(args, activityGroupID)
	}

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]*entity.Todo, 0)
	for rows.Next() {
		var t = new(entity.Todo)
		err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
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

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := "UPDATE todos SET title=?, priority=?, is_active=?, updated_at=? WHERE todo_id=?"

	// prepare the statement
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	args := []interface{}{
		todo.Title,
		todo.Priority,
		todo.IsActive,
		todo.UpdatedAt,
		todo.ID,
	}

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

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

	return nil
}
