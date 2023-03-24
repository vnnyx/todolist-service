// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	activity3 "github.com/vnnyx/golang-todo-api/internal/controller/activity"
	todo3 "github.com/vnnyx/golang-todo-api/internal/controller/todo"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/repository/activity"
	"github.com/vnnyx/golang-todo-api/internal/repository/todo"
	"github.com/vnnyx/golang-todo-api/internal/routes"
	activity2 "github.com/vnnyx/golang-todo-api/internal/usecase/activity"
	todo2 "github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

// Injectors from injector.go:

func InitializeRoute(configName string, e *fiber.App, c *cache.Cache) *routes.Route {
	config := infrastructure.NewConfig(configName)
	db := infrastructure.NewMySQLDatabase(config)
	activityRepository := activity.NewActivityRepository(db, c)
	activityUC := activity2.NewActivityUC(activityRepository)
	activityController := activity3.NewActivityController(activityUC, c)
	todoRepository := todo.NewTodoRepository(db, c)
	todoUC := todo2.NewTodoUC(todoRepository)
	todoController := todo3.NewTodoController(todoUC, c)
	route := routes.NewRoute(activityController, todoController, e)
	return route
}
