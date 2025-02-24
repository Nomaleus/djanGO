package storage

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrNoTasks    = errors.New("no tasks available")
	ErrTaskExists = errors.New("task already exists")
)
