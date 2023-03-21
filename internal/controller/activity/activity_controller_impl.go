package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityControllerImpl struct {
	activityUC activity.ActivityUC
	cache      *redis.Client
}

func NewActivityController(activityUC activity.ActivityUC, cache *redis.Client) ActivityController {
	return &ActivityControllerImpl{
		activityUC: activityUC,
		cache:      cache,
	}
}

func (controller *ActivityControllerImpl) InsertActivity(c echo.Context) error {
	var req web.ActivityCreateRequest
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	res, err := controller.activityUC.CreateActivity(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *ActivityControllerImpl) GetActivityByID(c echo.Context) error {
	var a *web.ActivityDTO
	var wg sync.WaitGroup

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	got, err := controller.cache.Get(c.Request().Context(), fmt.Sprintf("activity-%v", id)).Result()
	if err != nil {
		res, err := controller.activityUC.GetActivityByID(c.Request().Context(), int64(id))
		if err != nil {
			return err
		}

		data, err := json.Marshal(res)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = controller.cache.Set(c.Request().Context(), fmt.Sprintf("activity-%v", id), data, time.Until(time.Now().Add(time.Second*5))).Err()
			if err != nil {
				return
			}
		}()
		wg.Wait()

		return c.JSON(http.StatusOK, web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	err = json.Unmarshal([]byte(got), &a)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    a,
	})
}

func (controller *ActivityControllerImpl) GetAllActivity(c echo.Context) error {
	var a []*web.ActivityDTO
	var wg sync.WaitGroup

	got, err := controller.cache.Get(c.Request().Context(), "allactivity").Result()
	if err != nil {
		res, err := controller.activityUC.GetAllActivity(c.Request().Context())
		if err != nil {
			return err
		}

		data, err := json.Marshal(res)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = controller.cache.Set(c.Request().Context(), "allactivity", data, time.Until(time.Now().Add(time.Second*5))).Err()
			if err != nil {
				return
			}
		}()
		wg.Wait()

		return c.JSON(http.StatusOK, web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	err = json.Unmarshal([]byte(got), &a)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    a,
	})
}

func (controller *ActivityControllerImpl) UpdateActivity(c echo.Context) error {
	var req web.ActivityUpdateRequest
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	req.ID = int64(id)
	res, err := controller.activityUC.UpdateActivity(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *ActivityControllerImpl) DeleteActivity(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = controller.activityUC.DeleteActivity(c.Request().Context(), int64(id))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
