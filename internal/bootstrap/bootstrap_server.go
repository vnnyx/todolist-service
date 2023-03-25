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
	memdb := infrastructure.NewMemDB()

	activityRepo := activity.NewActivityRepository()
	err := activityRepo.InjectDB(mysqlDB)
	continueOrFatal(err)
	err = activityRepo.InjectCache(cache)
	continueOrFatal(err)
	err = activityRepo.InjectMemDB(memdb)
	continueOrFatal(err)

	activityUC := activityUC.NewActivityUC()
	err = activityUC.InjectActivityRepository(activityRepo)
	continueOrFatal(err)

	activityController := activityController.NewActivityController()
	err = activityController.InjectActivityUC(activityUC)
	continueOrFatal(err)

	todoRepo := todo.NewTodoRepository()
	err = todoRepo.InjectDB(mysqlDB)
	continueOrFatal(err)
	err = todoRepo.InjectCache(cache)
	continueOrFatal(err)
	err = todoRepo.InjectMemDB(memdb)
	continueOrFatal(err)

	todoUC := todoUC.NewTodoUC()
	err = todoUC.InjectTodoRepository(todoRepo)
	continueOrFatal(err)

	todoController := todoController.NewTodoController()
	err = todoController.InjectTodoUC(todoUC)
	continueOrFatal(err)

	go activityRepo.Worker()
	go todoRepo.Worker()

	r := routes.NewRoute(activityController, todoController, app)
	r.InitRoute()

	err = app.Listen(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}

func continueOrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
