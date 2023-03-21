package todo

import (
	"database/sql"
	"fmt"

	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type TodoRepositoryImpl struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) TodoRepository {
	return &TodoRepositoryImpl{
		db: db,
	}
}

func (repo *TodoRepositoryImpl) InsertTodo(todo entity.Todo) (*entity.Todo, error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "INSERT INTO todos(activity_group_id, title, is_active) VALUES(?,?,?)"
	args := []interface{}{
		todo.ActivityGroupID,
		todo.Title,
		todo.IsActive,
	}
	result, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	t, err := repo.GetTodoByID(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (repo *TodoRepositoryImpl) GetTodoByID(id int64) (todo *entity.Todo, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "SELECT * FROM todos WHERE todo_id=?"
	rows, err := repo.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var t entity.Todo
		err := rows.Scan(&t.ID, &t.ActivityGroupID, &t.Title, &t.IsActive, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	return nil, fmt.Errorf("Todo with ID %v Not Found", id)
}

func (repo *TodoRepositoryImpl) GetAllTodo(activityGroupID int64) (todos []*entity.Todo, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	var rows *sql.Rows

	switch {
	case activityGroupID != 0:
		query := "SELECT * FROM todos WHERE activity_group_id=?"
		rows, err = repo.db.QueryContext(ctx, query, activityGroupID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	default:
		query := "SELECT * FROM todos"
		rows, err = repo.db.QueryContext(ctx, query)
		if err != nil {
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
	return todos, nil
}

func (repo *TodoRepositoryImpl) UpdateTodo(todo entity.Todo) (*entity.Todo, error) {
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
