package activity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityController interface {
	InsertActivity(c *fiber.Ctx) error
	GetActivityByID(c *fiber.Ctx) error
	GetAllActivity(c *fiber.Ctx) error
	UpdateActivity(c *fiber.Ctx) error
	DeleteActivity(c *fiber.Ctx) error

	//dependency injection
	InjectActivityUC(activityUC activity.ActivityUC) error
}
