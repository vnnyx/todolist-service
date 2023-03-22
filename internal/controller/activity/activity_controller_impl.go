package activity

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityControllerImpl struct {
	activityUC activity.ActivityUC
	cache      *cache.Cache
}

func NewActivityController(activityUC activity.ActivityUC, cache *cache.Cache) ActivityController {
	return &ActivityControllerImpl{
		activityUC: activityUC,
		cache:      cache,
	}
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

	return c.Status(fiber.StatusCreated).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    res,
	})
}

func (controller *ActivityControllerImpl) GetActivityByID(c *fiber.Ctx) error {
	var wg sync.WaitGroup

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	data, found := controller.cache.Get(fmt.Sprintf("activity-%v", id))
	if !found {
		res, err := controller.activityUC.GetActivityByID(c.Context(), int64(id))
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			controller.cache.Set(fmt.Sprintf("activity-%v", id), res, time.Until(time.Now().Add(time.Second*5)))
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

func (controller *ActivityControllerImpl) GetAllActivity(c *fiber.Ctx) error {
	var wg sync.WaitGroup

	data, found := controller.cache.Get("allactivity")
	if !found {
		res, err := controller.activityUC.GetAllActivity(c.Context())
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			controller.cache.Set("allactivity", res, time.Until(time.Now().Add(time.Second*5)))
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

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    struct{}{},
	})
}
