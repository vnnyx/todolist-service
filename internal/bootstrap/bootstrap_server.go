package bootstrap

import (
	"log"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/exception"
	"github.com/vnnyx/golang-todo-api/internal/routes/di"
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
	c := cache.New(5*time.Minute, 10*time.Minute)
	r := di.InitializeRoute(".env", app, c)
	r.InitRoute()
	err := app.Listen(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}
