package rdb

import "errors"

var (
	ErrNotCreated = errors.New("not created")
	ErrNotDeleted = errors.New("not deleted")
	ErrNotFound   = errors.New("not found")
	ErrNotUpdated = errors.New("not updated")
)
