package domain

import "errors"

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryNameTaken = errors.New("category name already exists")

	ErrAssistantNotFound = errors.New("assistant not found")
	ErrAssistantInactive = errors.New("assistant is not active")

	ErrRunNotFound  = errors.New("run not found")
	ErrRunForbidden = errors.New("run belongs to another user")
)
