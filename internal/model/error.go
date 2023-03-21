package model

import "errors"

var (
	ErrTitleCannotBeNull           = errors.New("title cannot be null")
	ErrActivityGroupIDCannotBeNull = errors.New("activity_group_id cannot be null")
)
