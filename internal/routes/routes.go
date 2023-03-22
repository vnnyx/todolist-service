package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/controller/activity"
	"github.com/vnnyx/golang-todo-api/internal/controller/todo"
)

type Route struct {
	activityController activity.ActivityController
	todoController     todo.TodoController
	route              *fiber.App
}

func NewRoute(activityController activity.ActivityController, todoController todo.TodoController, route *fiber.App) *Route {
	return &Route{
		activityController: activityController,
		todoController:     todoController,
		route:              route,
	}
}

func (r *Route) InitRoute() {
	activity := r.route.Group("/activity-groups")
	activity.Post("", r.activityController.InsertActivity)
	activity.Get("/:id", r.activityController.GetActivityByID)
	activity.Get("", r.activityController.GetAllActivity)
	activity.Patch("/:id", r.activityController.UpdateActivity)
	activity.Delete("/:id", r.activityController.DeleteActivity)

	todo := r.route.Group("/todo-items")
	todo.Post("", r.todoController.InsertTodo)
	todo.Get("/:id", r.todoController.GetTodoByID)
	todo.Get("", r.todoController.GetAllTodo)
	todo.Patch("/:id", r.todoController.UpdateTodo)
	todo.Delete("/:id", r.todoController.DeleteTodo)
}
