package bootstrap

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vnnyx/golang-todo-api/internal/exception"
	"github.com/vnnyx/golang-todo-api/internal/routes/di"
)

func StarServer() {
	RunMigration()
	app := fiber.New(fiber.Config{
		ErrorHandler: exception.ErrorHandler,
	})
	app.Use(cors.New())
	app.Use(recover.New())
	r := di.InitializeRoute(".env", app)
	r.InitRoute()
	err := app.Listen(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}
