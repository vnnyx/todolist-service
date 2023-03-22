package todo

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/todo"
)

type TodoControllerImpl struct {
	todoUC todo.TodoUC
	cache  *cache.Cache
	lock   sync.Mutex
}

func NewTodoController(todoUC todo.TodoUC, cache *cache.Cache) TodoController {
	return &TodoControllerImpl{
		todoUC: todoUC,
		cache:  cache,
		lock:   sync.Mutex{},
	}
}

func (controller *TodoControllerImpl) InsertTodo(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
	controller.lock.Lock()
	defer controller.lock.Unlock()

	var wg sync.WaitGroup

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	data, found := controller.cache.Get(fmt.Sprintf("todo-%v", id))
	if !found {
		res, err := controller.todoUC.GetTodoByID(c.Context(), int64(id))
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			controller.cache.Set(fmt.Sprintf("todo-%v", id), res, time.Until(time.Now().Add(time.Second*5)))
		}()
		wg.Wait()
		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    data,
	})
}

func (controller *TodoControllerImpl) GetAllTodo(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

	var wg sync.WaitGroup

	activityGroupID, _ := strconv.Atoi(c.Query("activity_group_id"))
	data, found := controller.cache.Get(fmt.Sprintf("alltodo-%v", activityGroupID))
	if !found {
		res, err := controller.todoUC.GetAllTodo(c.Context(), int64(activityGroupID))
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			controller.cache.Set(fmt.Sprintf("alltodo-%v", activityGroupID), res, time.Until(time.Now().Add(time.Second*5)))
		}()
		wg.Wait()
		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    data,
	})
}

func (controller *TodoControllerImpl) UpdateTodo(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
