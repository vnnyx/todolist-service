package bootstrap

import (
	"context"
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

func StartServer() {
	RunMigration()

	// Set up the server and inject dependencies
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
	memdb := infrastructure.NewMemDB()

	activityRepo := activity.NewActivityRepository()
	if err := activityRepo.InjectDB(mysqlDB); err != nil {
		log.Fatalf("failed to inject DB into ActivityRepository: %v", err)
	}
	if err := activityRepo.InjectCache(cache); err != nil {
		log.Fatalf("failed to inject Cache into ActivityRepository: %v", err)
	}
	if err := activityRepo.InjectMemDB(memdb); err != nil {
		log.Fatalf("failed to inject MemDB into ActivityRepository: %v", err)
	}

	activityUC := activityUC.NewActivityUC()
	if err := activityUC.InjectActivityRepository(activityRepo); err != nil {
		log.Fatalf("failed to inject ActivityRepository into ActivityUC: %v", err)
	}

	activityController := activityController.NewActivityController()
	if err := activityController.InjectActivityUC(activityUC); err != nil {
		log.Fatalf("failed to inject ActivityUC into ActivityController: %v", err)
	}
	if err := activityController.InjectCache(cache); err != nil {
		log.Fatalf("failed to inject InjectCache into ActivityController: %v", err)
	}

	todoRepo := todo.NewTodoRepository()
	if err := todoRepo.InjectDB(mysqlDB); err != nil {
		log.Fatalf("failed to inject DB into TodoRepository: %v", err)
	}
	if err := todoRepo.InjectCache(cache); err != nil {
		log.Fatalf("failed to inject Cache into TodoRepository: %v", err)
	}
	if err := todoRepo.InjectMemDB(memdb); err != nil {
		log.Fatalf("failed to inject MemDB into TodoRepository: %v", err)
	}

	todoUC := todoUC.NewTodoUC()
	if err := todoUC.InjectTodoRepository(todoRepo); err != nil {
		log.Fatalf("failed to inject TodoRepository into TodoUC: %v", err)
	}

	todoController := todoController.NewTodoController()
	if err := todoController.InjectTodoUC(todoUC); err != nil {
		log.Fatalf("failed to inject TodoUC into TodoController: %v", err)
	}
	if err := todoController.InjectCache(cache); err != nil {
		log.Fatalf("failed to inject InjectCache into TodoController: %v", err)
	}

	// Start the workers for both repositories
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go activityRepo.Worker(ctx)
	go todoRepo.Worker(ctx)

	// Set up routes and start listening for requests
	r := routes.NewRoute(activityController, todoController, app)
	r.InitRoute()

	err := app.Listen(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}
