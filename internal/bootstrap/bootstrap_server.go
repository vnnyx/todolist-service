package bootstrap

import (
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patrickmn/go-cache"
	activityController "github.com/vnnyx/golang-todo-api/internal/controller/activity"
	todoController "github.com/vnnyx/golang-todo-api/internal/controller/todo"
	"github.com/vnnyx/golang-todo-api/internal/exception"
	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/repository/activity"
	"github.com/vnnyx/golang-todo-api/internal/repository/todo"
	"github.com/vnnyx/golang-todo-api/internal/routes"
	activityUC "github.com/vnnyx/golang-todo-api/internal/usecase/activity"
	todoUC "github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

func StarServer() {
	RunMigration()
	app := fiber.New(fiber.Config{
		ErrorHandler: exception.ErrorHandler,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})
	app.Use(cors.New())
	app.Use(recover.New())

	cfg := infrastructure.NewConfig(".env")
	cache := cache.New(5*time.Minute, 10*time.Minute)
	mysqlDB := infrastructure.NewMySQLDatabase(cfg)

	activityRepo := activity.NewActivityRepository()
	activityRepo.InjectDB(mysqlDB)
	activityRepo.InjectCache(cache)

	activityUC := activityUC.NewActivityUC()
	activityUC.InjectActivityRepository(activityRepo)

	activityController := activityController.NewActivityController()
	activityController.InjectActivityUC(activityUC)

	todoRepo := todo.NewTodoRepository()
	todoRepo.InjectDB(mysqlDB)
	todoRepo.InjectCache(cache)

	todoUC := todoUC.NewTodoUC()
	todoUC.InjectTodoRepository(todoRepo)

	todoController := todoController.NewTodoController()
	todoController.InjectTodoUC(todoUC)

	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				todoRepo.Worker()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				activityRepo.Worker()
			}
		}
	}()

	r := routes.NewRoute(activityController, todoController, app)
	r.InitRoute()

	err := app.Listen(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}
