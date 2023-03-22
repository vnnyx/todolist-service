package activity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/usecase/activity"
)

type ActivityControllerImpl struct {
	activityUC activity.ActivityUC
	cache      *redis.Client
	lock       sync.Mutex
}

func NewActivityController(activityUC activity.ActivityUC, cache *redis.Client) ActivityController {
	return &ActivityControllerImpl{
		activityUC: activityUC,
		cache:      cache,
		lock:       sync.Mutex{},
	}
}

func (controller *ActivityControllerImpl) InsertActivity(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
	controller.lock.Lock()
	defer controller.lock.Unlock()

	var a *web.ActivityDTO
	var wg sync.WaitGroup

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	got, err := controller.cache.Get(c.Context(), fmt.Sprintf("activity-%v", id)).Result()
	if err != nil {
		res, err := controller.activityUC.GetActivityByID(c.Context(), int64(id))
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
			err = controller.cache.Set(c.Context(), fmt.Sprintf("activity-%v", id), data, time.Until(time.Now().Add(time.Second*5))).Err()
			if err != nil {
				return
			}
		}()
		wg.Wait()

		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	err = json.Unmarshal([]byte(got), &a)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    a,
	})
}

func (controller *ActivityControllerImpl) GetAllActivity(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

	var a []*web.ActivityDTO
	var wg sync.WaitGroup

	got, err := controller.cache.Get(c.Context(), "allactivity").Result()
	if err != nil {
		res, err := controller.activityUC.GetAllActivity(c.Context())
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
			err = controller.cache.Set(c.Context(), "allactivity", data, time.Until(time.Now().Add(time.Second*5))).Err()
			if err != nil {
				return
			}
		}()
		wg.Wait()

		return c.Status(fiber.StatusOK).JSON(web.WebResponse{
			Status:  "Success",
			Message: "Success",
			Data:    res,
		})
	}

	err = json.Unmarshal([]byte(got), &a)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(web.WebResponse{
		Status:  "Success",
		Message: "Success",
		Data:    a,
	})
}

func (controller *ActivityControllerImpl) UpdateActivity(c *fiber.Ctx) error {
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
	controller.lock.Lock()
	defer controller.lock.Unlock()

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
