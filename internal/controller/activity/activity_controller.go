package activity

import "github.com/labstack/echo/v4"

type ActivityController interface {
	InsertActivity(c echo.Context) error
	GetActivityByID(c echo.Context) error
	GetAllActivity(c echo.Context) error
	UpdateActivity(c echo.Context) error
	DeleteActivity(c echo.Context) error
}
