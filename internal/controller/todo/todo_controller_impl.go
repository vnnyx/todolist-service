package todo

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

type TodoControllerImpl struct {
	todoUC todo.TodoUC
	cache  *cache.Cache
	mutex  sync.RWMutex
}

func NewTodoController() TodoController {
	return &TodoControllerImpl{
		mutex: sync.RWMutex{},
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (controller *TodoControllerImpl) InjectTodoUC(todoUC todo.TodoUC) error {
	controller.todoUC = todoUC
	return nil
}

func (controller *TodoControllerImpl) InjectCache(cache *cache.Cache) error {
	controller.cache = cache
	return nil
}

func (controller *TodoControllerImpl) InsertTodo(c *fiber.Ctx) error {
	var req web.TodoCreateRequest
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}

	res, err := controller.todoUC.CreateTodo(c.Context(), &req)
	if err != nil {
		return err
	}

	controller.cache.Flush()

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

	key := fmt.Sprintf("cotrollertodo-%d", id)

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	data, found := controller.cache.Get(key)
	if found {
		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    data,
		})
	}

	res, err := controller.todoUC.GetTodoByID(c.Context(), int64(id))
	if err != nil {
		return err
	}

	controller.cache.SetDefault(key, res)

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) GetAllTodo(c *fiber.Ctx) error {
	activityGroupID, _ := strconv.Atoi(c.Query("activity_group_id"))

	key := fmt.Sprintf("cotrolleralltodo-%d", activityGroupID)

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	data, found := controller.cache.Get(key)
	if found {
		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    data,
		})
	}

	res, err := controller.todoUC.GetAllTodo(c.Context(), int64(activityGroupID))
	if err != nil {
		return err
	}

	controller.cache.SetDefault(key, res)

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) UpdateTodo(c *fiber.Ctx) error {
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
	res, err := controller.todoUC.UpdateTodo(c.Context(), &req)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("cotrollertodo-%d", id)
	controller.mutex.Lock()
	controller.cache.SetDefault(key, res)
	controller.mutex.Unlock()

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *TodoControllerImpl) DeleteTodo(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	err = controller.todoUC.DeleteTodo(c.Context(), int64(id))
	if err != nil {
		return err
	}

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	key := fmt.Sprintf("cotrollertodo-%d", id)

	controller.cache.Delete(key)

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
