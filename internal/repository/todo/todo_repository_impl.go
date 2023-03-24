package todo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type TodoRepositoryImpl struct {
	db         *sql.DB
	cache      *cache.Cache
	workerTodo chan entity.Todo
}

func NewTodoRepository() TodoRepository {
	return &TodoRepositoryImpl{
		workerTodo: make(chan entity.Todo),
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

func (repo *TodoRepositoryImpl) Worker() {
	for {
		todo := <-repo.workerTodo
		query := "INSERT INTO todos(todo_id, activity_group_id, title, is_active, priority, created_at, updated_at) VALUES(?,?,?,?,?,?,?)"
		args := []interface{}{
			todo.ID,
			todo.ActivityGroupID,
			todo.Title,
			todo.IsActive,
			todo.Priority,
			todo.CreatedAt,
			todo.UpdatedAt,
		}
		_, err := repo.db.ExecContext(context.Background(), query, args...)
		if err != nil {
			return
		}
	}
}

func (repo *TodoRepositoryImpl) InsertTodo(todo entity.Todo) (*entity.Todo, error) {
	repo.cache.Flush()

	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	todo.ID = entity.TodoSeq
	entity.TodoSeq++

	go func(todo entity.Todo) {
		repo.workerTodo <- todo
	}(todo)

	return &todo, nil
}

func (repo *TodoRepositoryImpl) GetTodoByID(id int64) (todo *entity.Todo, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	data, found := repo.cache.Get(fmt.Sprintf("todoId-%v", id))
	if !found {
		query := "SELECT * FROM todos WHERE todo_id=?"
		rows, err := repo.db.QueryContext(ctx, query, id)
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
			repo.cache.SetDefault(fmt.Sprintf("todoId-%v", id), t)
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

	data, found := repo.cache.Get(fmt.Sprintf("alltodo-%v", activityGroupID))
	if !found {
		query := "SELECT * FROM todos"
		if activityGroupID != 0 {
			query = fmt.Sprintf("%s %s '%d'", query, " where activity_group_id = ", activityGroupID)
		}

		rows, err = repo.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var t = new(entity.Todo)
			err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
			if err != nil {
				return nil, err
			}
			todos = append(todos, t)
		}
		repo.cache.SetDefault(fmt.Sprintf("alltodo-%v", activityGroupID), todos)
		return todos, nil
	}

	return data.([]*entity.Todo), nil
}

func (repo *TodoRepositoryImpl) UpdateTodo(todo entity.Todo) (*entity.Todo, error) {
	repo.cache.Flush()

	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE todos SET title=?, priority=?, is_active=? WHERE todo_id=?"
	args := []interface{}{
		todo.Title,
		todo.Priority,
		todo.IsActive,
		todo.ID,
	}
	_, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	t, err := repo.GetTodoByID(todo.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
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
