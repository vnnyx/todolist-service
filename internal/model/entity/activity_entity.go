package entity

import (
	"time"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

type Activity struct {
	ID        int64 `gorm:"column:activity_id;primaryKey"`
	Title     string
	Email     string
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Activity) TableName() string {
	return "activities"
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
