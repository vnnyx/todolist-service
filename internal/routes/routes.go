package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/vnnyx/golang-todo-api/internal/controller/activity"
	"github.com/vnnyx/golang-todo-api/internal/controller/todo"
)

type Route struct {
	activityController activity.ActivityController
	todoController     todo.TodoController
	route              *echo.Echo
}

func NewRoute(activityController activity.ActivityController, todoController todo.TodoController, route *echo.Echo) *Route {
	return &Route{
		activityController: activityController,
		todoController:     todoController,
		route:              route,
	}
}

func (r *Route) InitRoute() {
	activity := r.route.Group("/activity-groups")
	activity.POST("", r.activityController.InsertActivity)
	activity.GET("/:id", r.activityController.GetActivityByID)
	activity.GET("", r.activityController.GetAllActivity)
	activity.PATCH("/:id", r.activityController.UpdateActivity)
	activity.DELETE("/:id", r.activityController.DeleteActivity)

	todo := r.route.Group("/todo-items")
	todo.POST("", r.todoController.InsertTodo)
	todo.GET("/:id", r.todoController.GetTodoByID)
	todo.GET("", r.todoController.GetAllTodo)
	todo.PATCH("/:id", r.todoController.UpdateTodo)
	todo.DELETE("/:id", r.todoController.DeleteTodo)
}
