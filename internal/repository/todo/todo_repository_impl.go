package todo

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

type TodoRepositoryImpl struct {
	db    *sql.DB
	cache *cache.Cache
	memdb *memdb.MemDB
}

func NewTodoRepository(db *sql.DB, cache *cache.Cache, memdb *memdb.MemDB) TodoRepository {
	return &TodoRepositoryImpl{
		db:    db,
		cache: cache,
		memdb: memdb,
	}
}

func (repo *TodoRepositoryImpl) InsertTodo(todo entity.Todo) (*entity.Todo, error) {
	var wg sync.WaitGroup
	var id = make(chan int64, 1)

	wg.Add(3)
	go func() {
		defer wg.Done()
		repo.cache.Flush()
	}()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	todo.Priority = "very-high"

	go func(todo entity.Todo) {
		defer wg.Done()
		query := "INSERT INTO todos(activity_group_id, title, is_active, created_at, updated_at) VALUES(?,?,?,?,?)"
		args := []interface{}{
			todo.ActivityGroupID,
			todo.Title,
			todo.IsActive,
			todo.CreatedAt,
			todo.UpdatedAt,
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
	}(todo)
	todo.ID = <-id
	go func(todo entity.Todo) {
		defer wg.Done()
		txn := repo.memdb.Txn(true)
		defer txn.Abort()
		err := txn.Insert("todos", todo)
		if err != nil {
			logrus.Error(err)
			return
		}
		txn.Commit()
	}(todo)
	wg.Wait()
	return &todo, nil
}

func (repo *TodoRepositoryImpl) GetTodoByID(id int64) (todo *entity.Todo, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	data, found := repo.cache.Get(fmt.Sprintf("id-%v", id))
	if !found {
		txn := repo.memdb.Txn(false)
		defer txn.Abort()
		raw, err := txn.First("todos", "id", id)
		if err != nil {
			return nil, err
		}
		if raw != nil {
			t := raw.(entity.Todo)
			return &t, nil
		}
		query := "SELECT * FROM todos WHERE todo_id=?"
		rows, err := repo.db.QueryContext(ctx, query, id)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			var t = new(entity.Todo)
			err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			repo.cache.Set(fmt.Sprintf("id-%v", id), t, cache.DefaultExpiration)
			return t, nil
		}
		return nil, fmt.Errorf("Todo with ID %v Not Found", id)
	}
	return data.(*entity.Todo), nil
}

func (repo *TodoRepositoryImpl) GetAllTodo(activityGroupID int64) (todos []*entity.Todo, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	var rows *sql.Rows
	var it memdb.ResultIterator

	data, found := repo.cache.Get(fmt.Sprintf("alltodo-%v", activityGroupID))
	if !found {
		switch {
		case activityGroupID != 0:
			txn := repo.memdb.Txn(false)
			defer txn.Abort()
			it, err = txn.Get("todos", "activity_group_id", activityGroupID)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			break
		default:
			txn := repo.memdb.Txn(false)
			defer txn.Abort()
			it, err = txn.Get("todos", "id")
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			break
		}
		for obj := it.Next(); obj != nil; obj = it.Next() {
			t := obj.(entity.Todo)
			todos = append(todos, &t)
		}
		if len(todos) > 0 {
			repo.cache.Set(fmt.Sprintf("alltodo-%v", activityGroupID), todos, cache.DefaultExpiration)
			return todos, nil
		}

		switch {
		case activityGroupID != 0:
			query := "SELECT * FROM todos WHERE activity_group_id=?"
			rows, err = repo.db.QueryContext(ctx, query, activityGroupID)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			defer rows.Close()
			break
		default:
			query := "SELECT * FROM todos"
			rows, err = repo.db.QueryContext(ctx, query)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			defer rows.Close()
		}
		for rows.Next() {
			var t entity.Todo
			err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
			if err != nil {
				return nil, err
			}
			todos = append(todos, &t)
		}
		repo.cache.Set(fmt.Sprintf("alltodo-%v", activityGroupID), todos, cache.DefaultExpiration)
		return todos, nil
	}
	return data.([]*entity.Todo), nil
}

func (repo *TodoRepositoryImpl) UpdateTodo(todo entity.Todo) (*entity.Todo, error) {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	todo.UpdatedAt = time.Now()

	query := "UPDATE todos SET title=?, priority=?, is_active=?, updated_at=? WHERE todo_id=?"
	args := []interface{}{
		todo.Title,
		todo.Priority,
		todo.IsActive,
		todo.UpdatedAt,
		todo.ID,
	}
	_, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (repo *TodoRepositoryImpl) DeleteTodo(id int64, title string) error {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "DELETE FROM todos WHERE todo_id=? AND title=?"
	args := []interface{}{
		id,
		title,
	}
	result, err := repo.db.ExecContext(ctx, query, args...)
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("Todo with ID %v and Title: %v Not Found", id, title)
	}
	if err != nil {
		return err
	}
	return nil
}
