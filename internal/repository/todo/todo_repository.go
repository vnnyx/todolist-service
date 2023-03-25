package todo

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-memdb"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type TodoRepository interface {
	InsertTodo(todo *entity.Todo) error
	GetTodoByID(id int64) (todo *entity.Todo, err error)
	GetAllTodo(activityGroupID int64) (todos []*entity.Todo, err error)
	UpdateTodo(todo *entity.Todo) error
	DeleteTodo(id int64, title string) error
	Worker(ctx context.Context)

	//dependency injection
	InjectDB(db *sql.DB) error
	InjectMemDB(memdb *memdb.MemDB) error
	InjectCache(cache *cache.Cache) error
}
