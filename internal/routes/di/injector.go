//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	activityController "github.com/vnnyx/golang-todo-api/internal/controller/activity"
	todoController "github.com/vnnyx/golang-todo-api/internal/controller/todo"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	activityRepo "github.com/vnnyx/golang-todo-api/internal/repository/activity"
	todoRepo "github.com/vnnyx/golang-todo-api/internal/repository/todo"
	"github.com/vnnyx/golang-todo-api/internal/routes"
	activityUC "github.com/vnnyx/golang-todo-api/internal/usecase/activity"
	todoUC "github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

func InitializeRoute(configName string, e *echo.Echo) *routes.Route {
	wire.Build(
		infrastructure.NewConfig,
		infrastructure.NewRedisClient,
		infrastructure.NewMySQLDatabase,
		activityRepo.NewActivityRepository,
		todoRepo.NewTodoRepository,
		activityUC.NewActivityUC,
		todoUC.NewTodoUC,
		activityController.NewActivityController,
		todoController.NewTodoController,
		routes.NewRoute,
	)
	return nil
}
