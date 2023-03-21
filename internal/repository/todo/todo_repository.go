package todo

import "github.com/vnnyx/golang-todo-api/internal/model/entity"

type TodoRepository interface {
	InsertTodo(todo entity.Todo) (*entity.Todo, error)
	GetTodoByID(id int64) (todo *entity.Todo, err error)
	GetAllTodo(activityGroupID int64) (todos []*entity.Todo, err error)
	UpdateTodo(todo entity.Todo) (*entity.Todo, error)
	DeleteTodo(id int64, title string) error
}
