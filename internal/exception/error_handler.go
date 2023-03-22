package exception

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/model"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	if databaseError(c, err) {
		return nil
	}
	generalError(c, err)
	return nil
}

func databaseError(c *fiber.Ctx, err error) bool {
	sqlError, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	switch {
	case sqlError.Number == 1062 && strings.Contains(sqlError.Message, "email"):
		_ = c.Status(fiber.StatusBadRequest).JSON(web.WebResponse{
			Status:  "Bad Request",
			Message: sqlError.Message,
		})
	default:
		_ = c.Status(fiber.StatusInternalServerError).JSON(web.WebResponse{
			Status:  "Internal Server Error",
			Message: sqlError.Message,
		})
	}
	return true
}

func generalError(c *fiber.Ctx, err error) {
	switch {
	case strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "Failed to Delete"):
		_ = c.Status(fiber.StatusNotFound).JSON(web.WebResponse{
			Status:  "Not Found",
			Message: err.Error(),
		})
	case errors.Is(err, model.ErrTitleCannotBeNull) || errors.Is(err, model.ErrActivityGroupIDCannotBeNull):
		_ = c.Status(fiber.StatusBadRequest).JSON(web.WebResponse{
			Status:  "Bad Request",
			Message: err.Error(),
		})
	default:
		_ = c.Status(fiber.StatusInternalServerError).JSON(web.WebResponse{
			Status:  "Internal Server Error",
			Message: err.Error(),
		})
	}
}
