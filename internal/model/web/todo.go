package web

import "time"

type TodoDTO struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	ActivityGroupID int64     `json:"activity_group_id"`
	IsActive        bool      `json:"is_active"`
	Priority        string    `json:"priority"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type TodoCreateRequest struct {
	Title           string `json:"title"`
	ActivityGroupID int64  `json:"activity_group_id"`
	IsActive        *bool  `json:"is_active"`
}

type TodoUpdateRequest struct {
	ID       int64
	Title    string `json:"title"`
	Priority string `json:"priority"`
	IsActive *bool  `json:"is_active"`
	Status   string `json:"status"`
}
