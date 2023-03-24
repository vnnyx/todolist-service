package entity

import (
	"time"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

var (
	TodoSeq = int64(1)
)

type Todo struct {
	ID              int64
	ActivityGroupID int64
	Title           string
	IsActive        bool
	Priority        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (t Todo) ToDTO() *web.TodoDTO {
	return &web.TodoDTO{
		ID:              t.ID,
		Title:           t.Title,
		ActivityGroupID: t.ActivityGroupID,
		IsActive:        t.IsActive,
		Priority:        t.Priority,
		CreatedAt:       t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       t.UpdatedAt.Format(time.RFC3339),
	}
}
