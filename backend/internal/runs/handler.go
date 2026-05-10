package runs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/httpx"
	"avito-internship-fs/internal/llm"
)

const maxUserPromptLen = 8000

type Service interface {
	Run(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("assistantId")
	assistantID, err := uuid.Parse(rawID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid assistantId")
		return
	}

	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
		return
	}

	var req api.PostAssistantsAssistantIdRunJSONRequestBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "malformed request body")
		return
	}
	prompt := strings.TrimSpace(req.UserPrompt)
	if prompt == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "userPrompt is required")
		return
	}
	if len(prompt) > maxUserPromptLen {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "userPrompt is too long")
		return
	}

	run, err := h.svc.Run(r.Context(), assistantID, principal.UserID, prompt)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAssistantNotFound):
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeAssistantNotFound, "assistant not found")
		case errors.Is(err, domain.ErrAssistantInactive):
			httpx.WriteError(w, http.StatusConflict, httpx.CodeAssistantInactive, "assistant is not active")
		case errors.Is(err, llm.ErrProviderFailed):
			httpx.WriteError(w, http.StatusBadGateway, httpx.CodeLLMProviderError, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to run assistant")
		}

		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toAPIRun(run))
}

func toAPIRun(r domain.AssistantRun) api.AssistantRun {
	return api.AssistantRun{
		Id:          r.ID,
		AssistantId: r.AssistantID,
		UserId:      r.UserID,
		Model:       r.Model,
		UserPrompt:  r.UserPrompt,
		Output:      r.Output,
		Status:      api.RunStatus(r.Status),
		Error:       r.Error,
		CreatedAt:   new(r.CreatedAt.UTC()),
	}
}
