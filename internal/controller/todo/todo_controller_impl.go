package todo

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
	"github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

type TodoControllerImpl struct {
	todoUC todo.TodoUC
	cache  *redis.Client
	lock   sync.Mutex
}

func NewTodoController(todoUC todo.TodoUC, cache *redis.Client) TodoController {
	return &TodoControllerImpl{
		todoUC: todoUC,
		cache:  cache,
		lock:   sync.Mutex{},
	}
}

func (controller *TodoControllerImpl) InsertTodo(c echo.Context) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()
	var req web.TodoCreateRequest
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	res, err := controller.todoUC.CreateTodo(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) GetTodoByID(c echo.Context) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()
	var t *web.TodoDTO
	var wg sync.WaitGroup

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	got, err := controller.cache.Get(c.Request().Context(), fmt.Sprintf("todo-%v", id)).Result()
	if err != nil {
		res, err := controller.todoUC.GetTodoByID(c.Request().Context(), int64(id))
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
			err = controller.cache.Set(c.Request().Context(), fmt.Sprintf("todo-%v", id), data, time.Until(time.Now().Add(time.Second*5))).Err()
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

	err = json.Unmarshal([]byte(got), &t)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    t,
	})
}

func (controller *TodoControllerImpl) GetAllTodo(c echo.Context) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()
	var t []*web.TodoDTO
	var wg sync.WaitGroup

	activityGroupID, _ := strconv.Atoi(c.QueryParam("activity_group_id"))
	got, err := controller.cache.Get(c.Request().Context(), fmt.Sprintf("alltodo-%v", activityGroupID)).Result()
	if err != nil {
		res, err := controller.todoUC.GetAllTodo(c.Request().Context(), int64(activityGroupID))
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
			err = controller.cache.Set(c.Request().Context(), fmt.Sprintf("alltodo-%v", activityGroupID), data, time.Until(time.Now().Add(time.Second*5))).Err()
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

	err = json.Unmarshal([]byte(got), &t)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    t,
	})
}

func (controller *TodoControllerImpl) UpdateTodo(c echo.Context) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()
	var req web.TodoUpdateRequest
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	req.ID = int64(id)
	res, err := controller.todoUC.UpdateTodo(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) DeleteTodo(c echo.Context) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()
	var req web.TodoDeleteRequest
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	req.ID = int64(id)
	err = controller.todoUC.DeleteTodo(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
