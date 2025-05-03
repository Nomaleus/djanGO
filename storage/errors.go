package storage

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrTaskExists = errors.New("task already exists")
)
