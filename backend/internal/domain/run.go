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

type RunFeedbackRating int

const (
	RunFeedbackDislike RunFeedbackRating = -1
	RunFeedbackLike    RunFeedbackRating = 1
)

type AssistantRun struct {
	ID             uuid.UUID
	AssistantID    uuid.UUID
	AssistantName  *string
	CategoryID     *uuid.UUID
	CategoryName   *string
	UserID         uuid.UUID
	Model          string
	UserPrompt     string
	Output         *string
	Status         RunStatus
	Error          *string
	TokensIn       *int
	TokensOut      *int
	LatencyMs      *int
	FinishReason   *string
	FeedbackRating *RunFeedbackRating
	CreatedAt      time.Time
}
