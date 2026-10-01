package common

import "errors"

var (
	ErrNotFound        = errors.New("resource not found")
	ErrDuplicateDomain = errors.New("website domain already exists")
)
