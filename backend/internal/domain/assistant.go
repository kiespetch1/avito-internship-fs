package domain

import (
	"time"

	"github.com/google/uuid"
)

type Assistant struct {
	ID                uuid.UUID
	CategoryID        uuid.UUID
	CategoryName      *string
	Name              string
	Description       string
	Model             string
	SystemPrompt      string
	ExampleUserPrompt *string
	IsActive          bool
	IsFavorite        bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
