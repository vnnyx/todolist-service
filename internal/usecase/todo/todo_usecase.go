package todo

import (
	"context"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

type TodoUC interface {
	CreateTodo(ctx context.Context, req web.TodoCreateRequest) (*web.TodoDTO, error)
	GetTodoByID(ctx context.Context, id int64) (*web.TodoDTO, error)
	GetAllTodo(ctx context.Context, activityGroupID int64) ([]*web.TodoDTO, error)
	UpdateTodo(ctx context.Context, req web.TodoUpdateRequest) (*web.TodoDTO, error)
	DeleteTodo(ctx context.Context, req web.TodoDeleteRequest) error
}
