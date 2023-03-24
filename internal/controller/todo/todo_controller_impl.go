package todo

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/controller"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

type TodoControllerImpl struct {
	todoUC todo.TodoUC
	cache  *controller.LocalCache
}

func NewTodoController(todoUC todo.TodoUC) TodoController {
	return &TodoControllerImpl{
		todoUC: todoUC,
		cache: &controller.LocalCache{
			Cache: make(map[string]interface{}),
			Mu:    sync.Mutex{},
		},
	}
}

func (controller *TodoControllerImpl) InsertTodo(c *fiber.Ctx) error {
	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	var req web.TodoCreateRequest
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}

	res, err := controller.todoUC.CreateTodo(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) GetTodoByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	controller.cache.Mu.Lock()
	cachedData, found := controller.cache.Cache[fmt.Sprintf("todo-%v", id)]
	controller.cache.Mu.Unlock()
	if !found {
		res, err := controller.todoUC.GetTodoByID(c.Context(), int64(id))
		if err != nil {
			return err
		}
		controller.cache.Mu.Lock()
		controller.cache.Cache[fmt.Sprintf("todo-%v", id)] = res
		controller.cache.Mu.Unlock()

		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    cachedData,
	})
}

func (controller *TodoControllerImpl) GetAllTodo(c *fiber.Ctx) error {
	activityGroupID, _ := strconv.Atoi(c.Query("activity_group_id"))

	controller.cache.Mu.Lock()
	cachedData, found := controller.cache.Cache[fmt.Sprintf("alltodo-%v", activityGroupID)]
	controller.cache.Mu.Unlock()
	if !found {
		res, err := controller.todoUC.GetAllTodo(c.Context(), int64(activityGroupID))
		if err != nil {
			return err
		}
		controller.cache.Mu.Lock()
		controller.cache.Cache[fmt.Sprintf("alltodo-%v", activityGroupID)] = res
		controller.cache.Mu.Unlock()
		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    cachedData,
	})
}

func (controller *TodoControllerImpl) UpdateTodo(c *fiber.Ctx) error {
	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	var req web.TodoUpdateRequest
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	err = c.BodyParser(&req)
	if err != nil {
		return err
	}
	req.ID = int64(id)
	res, err := controller.todoUC.UpdateTodo(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) DeleteTodo(c *fiber.Ctx) error {
	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	err = controller.todoUC.DeleteTodo(c.Context(), int64(id))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
