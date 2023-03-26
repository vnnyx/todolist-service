package activity

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityControllerImpl struct {
	activityUC activity.ActivityUC
	cache      *cache.Cache
	mutex      sync.RWMutex
}

func NewActivityController() ActivityController {
	return &ActivityControllerImpl{
		mutex: sync.RWMutex{},
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (controller *ActivityControllerImpl) InjectActivityUC(activityUC activity.ActivityUC) error {
	controller.activityUC = activityUC
	return nil
}

func (controller *ActivityControllerImpl) InjectCache(cache *cache.Cache) error {
	controller.cache = cache
	return nil
}

func (controller *ActivityControllerImpl) InsertActivity(c *fiber.Ctx) error {
	var req web.ActivityCreateRequest
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}
	res, err := controller.activityUC.CreateActivity(c.Context(), &req)
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

func (controller *ActivityControllerImpl) GetActivityByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	key := fmt.Sprintf("cotrolleractivity-%d", id)

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

	res, err := controller.activityUC.GetActivityByID(c.Context(), int64(id))
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

func (controller *ActivityControllerImpl) GetAllActivity(c *fiber.Ctx) error {
	key := "controllerallactivity"

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
	res, err := controller.activityUC.GetAllActivity(c.Context())
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
	res, err := controller.activityUC.UpdateActivity(c.Context(), &req)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("cotrolleractivity-%d", id)
	controller.mutex.Lock()
	controller.cache.SetDefault(key, res)
	controller.mutex.Unlock()

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

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	key := fmt.Sprintf("cotrolleractivity-%d", id)
	controller.cache.Delete(key)

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
