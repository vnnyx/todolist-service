package exception

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/vnnyx/golang-todo-api/internal/model"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

func ErrorHandler(err error, ctx echo.Context) {
	if databaseError(err, ctx) {
		return
	}
	generalError(err, ctx)
}

func databaseError(err error, ctx echo.Context) bool {
	sqlError, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	switch {
	case sqlError.Number == 1062 && strings.Contains(sqlError.Message, "email"):
		_ = ctx.JSON(http.StatusBadRequest, web.WebResponse{
			Status:  "Bad Request",
			Message: sqlError.Message,
		})
	default:
		_ = ctx.JSON(http.StatusInternalServerError, web.WebResponse{
			Status:  "Internal Server Error",
			Message: sqlError.Message,
		})
	}
	return true
}

func generalError(err error, ctx echo.Context) {
	switch {
	case strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "Failed to Delete"):
		_ = ctx.JSON(http.StatusNotFound, web.WebResponse{
			Status:  "Not Found",
			Message: err.Error(),
		})
	case errors.Is(err, model.ErrTitleCannotBeNull) || errors.Is(err, model.ErrActivityGroupIDCannotBeNull):
		_ = ctx.JSON(http.StatusBadRequest, web.WebResponse{
			Status:  "Bad Request",
			Message: err.Error(),
		})
	default:
		_ = ctx.JSON(http.StatusInternalServerError, web.WebResponse{
			Status:  "Internal Server Error",
			Message: err.Error(),
		})
	}
}
