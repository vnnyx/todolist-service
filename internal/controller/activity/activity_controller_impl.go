package activity

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/controller"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityControllerImpl struct {
	activityUC activity.ActivityUC
	cache      *controller.LocalCache
}

func NewActivityController() ActivityController {
	return &ActivityControllerImpl{
		cache: &controller.LocalCache{
			Cache: make(map[string]interface{}),
			Mu:    sync.Mutex{},
		},
	}
}

func (controller *ActivityControllerImpl) InjectActivityUC(activityUC activity.ActivityUC) error {
	controller.activityUC = activityUC
	return nil
}

func (controller *ActivityControllerImpl) InsertActivity(c *fiber.Ctx) error {
	var req web.ActivityCreateRequest
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}
	res, err := controller.activityUC.CreateActivity(c.Context(), req)
	if err != nil {
		return err
	}

	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	return c.Status(fiber.StatusCreated).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *ActivityControllerImpl) GetActivityByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	controller.cache.Mu.Lock()
	data, found := controller.cache.Cache[fmt.Sprintf("activityId-%v", id)]
	controller.cache.Mu.Unlock()
	if !found {
		res, err := controller.activityUC.GetActivityByID(c.Context(), int64(id))
		if err != nil {
			return err
		}

		go func(res *web.ActivityDTO) {
			controller.cache.Mu.Lock()
			controller.cache.Cache[fmt.Sprintf("activityId-%v", id)] = res
			controller.cache.Mu.Unlock()
		}(res)

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

func (controller *ActivityControllerImpl) GetAllActivity(c *fiber.Ctx) error {
	controller.cache.Mu.Lock()
	data, found := controller.cache.Cache["allactivity"]
	controller.cache.Mu.Unlock()
	if !found {
		res, err := controller.activityUC.GetAllActivity(c.Context())
		if err != nil {
			return err
		}

		go func(res []*web.ActivityDTO) {
			controller.cache.Mu.Lock()
			controller.cache.Cache["allactivity"] = res
			controller.cache.Mu.Unlock()
		}(res)

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

func (controller *ActivityControllerImpl) UpdateActivity(c *fiber.Ctx) error {
	var req web.ActivityUpdateRequest
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	err = c.BodyParser(&req)
	if err != nil {
		return err
	}
	req.ID = int64(id)
	res, err := controller.activityUC.UpdateActivity(c.Context(), req)
	if err != nil {
		return err
	}

	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *ActivityControllerImpl) DeleteActivity(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	err = controller.activityUC.DeleteActivity(c.Context(), int64(id))
	if err != nil {
		return err
	}

	go func() {
		controller.cache.Mu.Lock()
		controller.cache.Cache = make(map[string]interface{})
		controller.cache.Mu.Unlock()
	}()

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
