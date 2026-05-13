package runs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/httpx"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/service"
)

const maxUserPromptLen = 8000

type Service interface {
	Run(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error)
	RunStream(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error)
	List(ctx context.Context, in service.RunListInput) ([]domain.AssistantRun, int, error)
	SetFeedback(ctx context.Context, runID, userID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

type listResponse struct {
	Runs       []api.AssistantRun `json:"runs"`
	Pagination api.Pagination     `json:"pagination"`
}

type streamDeltaResponse struct {
	Delta string `json:"delta"`
}

type streamFailureResponse struct {
	Run   api.AssistantRun   `json:"run"`
	Error streamErrorPayload `json:"error"`
}

type streamErrorPayload struct {
	Code    httpx.ErrorCode `json:"code"`
	Message string          `json:"message"`
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
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	prompt, err := httpx.RequireField("userPrompt", req.UserPrompt, maxUserPromptLen)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
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

func (h *Handler) RunStream(w http.ResponseWriter, r *http.Request) {
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

	var req api.PostAssistantsAssistantIdRunStreamJSONRequestBody
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	prompt, err := httpx.RequireField("userPrompt", req.UserPrompt, maxUserPromptLen)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "streaming is not supported")
		return
	}

	streamCtx, cancelStream := context.WithCancel(r.Context())
	defer cancelStream()

	streamStarted := false
	var streamWriteErr error
	startStream := func() {
		if streamStarted {
			return
		}
		streamStarted = true
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
	}
	writeEvent := func(event string, payload any) {
		if streamWriteErr != nil || streamCtx.Err() != nil {
			return
		}
		startStream()
		if err := writeSSE(w, event, payload); err != nil {
			streamWriteErr = err
			cancelStream()

			return
		}
		flusher.Flush()
	}

	run, err := h.svc.RunStream(streamCtx, assistantID, principal.UserID, prompt, service.RunStreamCallbacks{
		OnRunCreated: func(run domain.AssistantRun) {
			writeEvent("run", toAPIRun(run))
		},
		OnDelta: func(delta string) {
			writeEvent("delta", streamDeltaResponse{Delta: delta})
		},
	})
	if err != nil {
		if streamStarted {
			if streamWriteErr != nil || errors.Is(streamCtx.Err(), context.Canceled) {
				return
			}
			writeEvent("failed", streamFailureResponse{
				Run: toAPIRun(run),
				Error: streamErrorPayload{
					Code:    streamErrorCode(err),
					Message: streamErrorMessage(err),
				},
			})

			return
		}

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

	writeEvent("done", toAPIRun(run))
}

func (h *Handler) MyRuns(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
		return
	}
	in := service.RunListInput{UserID: new(principal.UserID)}
	if err := applyRunListQuery(r, &in, false); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return
	}

	h.writeList(w, r.Context(), in)
}

func (h *Handler) AdminRuns(w http.ResponseWriter, r *http.Request) {
	in := service.RunListInput{}
	if err := applyRunListQuery(r, &in, true); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return
	}

	h.writeList(w, r.Context(), in)
}

func (h *Handler) SetFeedback(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("runId")
	runID, err := uuid.Parse(rawID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid runId")
		return
	}

	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
		return
	}

	var req api.PutRunsRunIdFeedbackJSONRequestBody
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}

	rating := domain.RunFeedbackRating(req.Rating)
	switch rating {
	case domain.RunFeedbackLike, domain.RunFeedbackDislike:
	default:
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "rating must be -1 or 1")
		return
	}

	run, err := h.svc.SetFeedback(r.Context(), runID, principal.UserID, rating)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRunNotFound):
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "run not found")
		case errors.Is(err, domain.ErrRunForbidden):
			httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "run belongs to another user")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to save feedback")
		}

		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAPIRun(run))
}

func (h *Handler) writeList(w http.ResponseWriter, ctx context.Context, in service.RunListInput) {
	items, total, err := h.svc.List(ctx, in)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to list runs")
		return
	}

	resp := listResponse{
		Runs:       make([]api.AssistantRun, 0, len(items)),
		Pagination: api.Pagination{Page: in.Page, PageSize: in.PageSize, Total: total},
	}
	for _, run := range items {
		resp.Runs = append(resp.Runs, toAPIRun(run))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func applyRunListQuery(r *http.Request, in *service.RunListInput, allowAssistantID bool) error {
	q := r.URL.Query()

	if raw := q.Get("status"); raw != "" {
		switch domain.RunStatus(raw) {
		case domain.RunPending, domain.RunSuccess, domain.RunFailed:
			in.Status = new(domain.RunStatus(raw))
		default:
			return errors.New("invalid status")
		}
	}
	if allowAssistantID {
		if raw := q.Get("assistantId"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				return errors.New("invalid assistantId")
			}
			in.AssistantID = &id
		}
	}
	page, pageSize, err := httpx.PageParams(q, 20)
	if err != nil {
		return err
	}
	in.Page, in.PageSize = page, pageSize

	return nil
}

func writeSSE(w http.ResponseWriter, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte("event: " + event + "\n")); err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n\n"))

	return err
}

func streamErrorCode(err error) httpx.ErrorCode {
	switch {
	case errors.Is(err, llm.ErrProviderFailed):
		return httpx.CodeLLMProviderError
	default:
		return httpx.CodeInternalError
	}
}

func streamErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func toAPIRun(r domain.AssistantRun) api.AssistantRun {
	out := api.AssistantRun{
		Id:            r.ID,
		AssistantId:   r.AssistantID,
		AssistantName: r.AssistantName,
		CategoryName:  r.CategoryName,
		UserId:        r.UserID,
		Model:         r.Model,
		UserPrompt:    r.UserPrompt,
		Output:        r.Output,
		Status:        api.RunStatus(r.Status),
		Error:         r.Error,
		CreatedAt:     new(r.CreatedAt.UTC()),
	}
	if r.CategoryID != nil {
		out.CategoryId = new(*r.CategoryID)
	}
	if r.FeedbackRating != nil {
		rating := api.RunFeedbackRating(*r.FeedbackRating)
		out.FeedbackRating = &rating
	}

	return out
}
