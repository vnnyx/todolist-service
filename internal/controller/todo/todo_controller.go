package todo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

type TodoController interface {
	InsertTodo(c *fiber.Ctx) error
	GetTodoByID(c *fiber.Ctx) error
	GetAllTodo(c *fiber.Ctx) error
	UpdateTodo(c *fiber.Ctx) error
	DeleteTodo(c *fiber.Ctx) error

	//dependency injection
	InjectTodoUC(todoUC todo.TodoUC) error
	InjectCache(cache *cache.Cache) error
}
