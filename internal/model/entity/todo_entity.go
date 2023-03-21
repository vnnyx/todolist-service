package entity

import (
	"time"

	"github.com/vnnyx/golang-todo-api/internal/model/web"
)

type Todo struct {
	ID              int64 `gorm:"column:todo_id;primaryKey"`
	ActivityGroupID int64
	Title           string
	IsActive        bool      `gorm:"default:true"`
	Priority        string    `gorm:"default:very-high"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Todo) TableName() string {
	return "todos"
}

func (t Todo) ToDTO() *web.TodoDTO {
	return &web.TodoDTO{
		ID:              t.ID,
		Title:           t.Title,
		ActivityGroupID: t.ActivityGroupID,
		IsActive:        t.IsActive,
		Priority:        t.Priority,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
