package todo

import "github.com/labstack/echo/v4"

type TodoController interface {
	InsertTodo(c echo.Context) error
	GetTodoByID(c echo.Context) error
	GetAllTodo(c echo.Context) error
	UpdateTodo(c echo.Context) error
	DeleteTodo(c echo.Context) error
}
