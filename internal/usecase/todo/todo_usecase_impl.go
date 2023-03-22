package todo

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/vnnyx/golang-todo-api/internal/model"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
	"github.com/vnnyx/golang-todo-api/internal/model/web"
	"github.com/vnnyx/golang-todo-api/internal/repository/todo"
)

type TodoUCImpl struct {
	todoRepository todo.TodoRepository
}

func NewTodoUC(todoRepository todo.TodoRepository) TodoUC {
	return &TodoUCImpl{
		todoRepository: todoRepository,
	}
}

func (uc *TodoUCImpl) CreateTodo(ctx context.Context, req web.TodoCreateRequest) (*web.TodoDTO, error) {
	if req.ActivityGroupID == 0 {
		return nil, model.ErrActivityGroupIDCannotBeNull
	}
	if req.Title == "" {
		return nil, model.ErrTitleCannotBeNull
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	got, err := uc.todoRepository.InsertTodo(entity.Todo{
		ActivityGroupID: req.ActivityGroupID,
		Title:           req.Title,
		IsActive:        isActive,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return got.ToDTO(), nil
}

func (uc *TodoUCImpl) GetTodoByID(ctx context.Context, id int64) (*web.TodoDTO, error) {
	got, err := uc.todoRepository.GetTodoByID(id)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return got.ToDTO(), nil
}

func (uc *TodoUCImpl) GetAllTodo(ctx context.Context, activityGroupID int64) ([]*web.TodoDTO, error) {
	got, err := uc.todoRepository.GetAllTodo(activityGroupID)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	res := make([]*web.TodoDTO, 0)
	for _, t := range got {
		res = append(res, t.ToDTO())
	}

	return res, nil
}

func (uc *TodoUCImpl) UpdateTodo(ctx context.Context, req web.TodoUpdateRequest) (*web.TodoDTO, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	todo, err := uc.todoRepository.GetTodoByID(req.ID)
	if err != nil {
		return nil, err
	}

	todo.IsActive = isActive
	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Priority != "" {
		todo.Priority = req.Priority
	}

	got, err := uc.todoRepository.UpdateTodo(*todo)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return got.ToDTO(), nil
}

func (uc *TodoUCImpl) DeleteTodo(ctx context.Context, id int64) error {
	todo, err := uc.todoRepository.GetTodoByID(id)
	if err != nil {
		return err
	}

	err = uc.todoRepository.DeleteTodo(todo.ID, todo.Title)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}
