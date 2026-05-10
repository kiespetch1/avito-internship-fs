package domain

import (
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	RunPending RunStatus = "pending"
	RunSuccess RunStatus = "success"
	RunFailed  RunStatus = "failed"
)

type AssistantRun struct {
	ID           uuid.UUID
	AssistantID  uuid.UUID
	UserID       uuid.UUID
	Model        string
	UserPrompt   string
	Output       *string
	Status       RunStatus
	Error        *string
	TokensIn     *int
	TokensOut    *int
	LatencyMs    *int
	FinishReason *string
	CreatedAt    time.Time
}
