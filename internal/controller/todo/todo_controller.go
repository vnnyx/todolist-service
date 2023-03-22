package todo

import (
	"github.com/gofiber/fiber/v2"
)

type TodoController interface {
	InsertTodo(c *fiber.Ctx) error
	GetTodoByID(c *fiber.Ctx) error
	GetAllTodo(c *fiber.Ctx) error
	UpdateTodo(c *fiber.Ctx) error
	DeleteTodo(c *fiber.Ctx) error
}
