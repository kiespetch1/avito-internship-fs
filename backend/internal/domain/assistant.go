package domain

import (
	"time"

	"github.com/google/uuid"
)

type Assistant struct {
	ID                uuid.UUID
	CategoryID        uuid.UUID
	Name              string
	Description       string
	Model             string
	SystemPrompt      string
	ExampleUserPrompt *string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
