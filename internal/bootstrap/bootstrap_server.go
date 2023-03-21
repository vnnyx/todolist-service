package bootstrap

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/vnnyx/golang-todo-api/internal/exception"
	"github.com/vnnyx/golang-todo-api/internal/routes/di"
)

func StarServer() {
	e := echo.New()
	e.HTTPErrorHandler = exception.ErrorHandler
	r := di.InitializeRoute(".env", e)
	r.InitRoute()
	err := e.Start(":3030")
	if err != nil {
		log.Fatalf("couldn't start server: %v", err)
	}
}
