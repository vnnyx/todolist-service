package entity

import (
	"time"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

type Activity struct {
	ID        int64
	Title     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a Activity) ToDTO() *web.ActivityDTO {
	return &web.ActivityDTO{
		ID:        a.ID,
		Title:     a.Title,
		Email:     a.Email,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
}
