package domain

import "errors"

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryNameTaken = errors.New("category name already exists")
)
