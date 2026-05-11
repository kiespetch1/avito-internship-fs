package assistants

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/httpx"
	"avito-internship-fs/internal/service"
)

const (
	maxAssistantNameLen          = 256
	maxAssistantDescriptionLen   = 2000
	maxAssistantModelLen         = 128
	maxAssistantSystemPromptLen  = 32_000
	maxAssistantExamplePromptLen = 8000
)

type Service interface {
	Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error)
	Create(ctx context.Context, in service.AssistantCreateInput) (domain.Assistant, error)
	Update(ctx context.Context, in service.AssistantUpdateInput) (domain.Assistant, error)
	List(ctx context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

type listResponse struct {
	Assistants []api.Assistant `json:"assistants"`
	Pagination api.Pagination  `json:"pagination"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
		return
	}

	q := r.URL.Query()
	in := service.AssistantListInput{}

	if raw := q.Get("categoryId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid categoryId")
			return
		}
		in.CategoryID = &id
	}
	if raw := strings.TrimSpace(q.Get("q")); raw != "" {
		in.Query = &raw
	}
	if raw := q.Get("includeInactive"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid includeInactive")
			return
		}
		if v && principal.Role != auth.RoleAdmin {
			httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "includeInactive is admin-only")
			return
		}
		in.IncludeInactive = v
	}
	page, pageSize, err := httpx.PageParams(q, 10)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return
	}
	in.Page, in.PageSize = page, pageSize

	items, total, err := h.svc.List(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to list assistants")
		return
	}

	out := listResponse{
		Assistants: make([]api.Assistant, 0, len(items)),
		Pagination: api.Pagination{Page: in.Page, PageSize: in.PageSize, Total: total},
	}
	hideSystem := principal.Role != auth.RoleAdmin
	for _, a := range items {
		out.Assistants = append(out.Assistants, toAPIAssistant(a, hideSystem))
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
		return
	}

	id, err := uuid.Parse(r.PathValue("assistantId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid assistantId")
		return
	}

	a, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeAssistantError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAPIAssistant(a, principal.Role != auth.RoleAdmin))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req api.PostAssistantsJSONRequestBody
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	f, ok := validateAssistantFields(w, req.Name, req.Description, req.Model, req.SystemPrompt, req.ExampleUserPrompt)
	if !ok {
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	a, err := h.svc.Create(r.Context(), service.AssistantCreateInput{
		CategoryID:        req.CategoryId,
		Name:              f.Name,
		Description:       f.Description,
		Model:             f.Model,
		SystemPrompt:      f.SystemPrompt,
		ExampleUserPrompt: f.ExamplePrompt,
		IsActive:          isActive,
	})
	if err != nil {
		writeAssistantError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toAPIAssistant(a, false))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("assistantId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "invalid assistantId")
		return
	}

	var req api.PutAssistantsAssistantIdJSONRequestBody
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	f, ok := validateAssistantFields(w, req.Name, req.Description, req.Model, req.SystemPrompt, req.ExampleUserPrompt)
	if !ok {
		return
	}

	a, err := h.svc.Update(r.Context(), service.AssistantUpdateInput{
		ID:                id,
		CategoryID:        req.CategoryId,
		Name:              f.Name,
		Description:       f.Description,
		Model:             f.Model,
		SystemPrompt:      f.SystemPrompt,
		ExampleUserPrompt: f.ExamplePrompt,
		IsActive:          req.IsActive,
	})
	if err != nil {
		writeAssistantError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAPIAssistant(a, false))
}

func writeAssistantError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAssistantNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeAssistantNotFound, "assistant not found")
	case errors.Is(err, domain.ErrCategoryNotFound):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeCategoryNotFound, "category not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
	}
}

type assistantFields struct {
	Name          string
	Description   string
	Model         string
	SystemPrompt  string
	ExamplePrompt *string
}

func validateAssistantFields(w http.ResponseWriter, rawName, rawDesc, rawModel, rawSystem string, rawExample *string) (assistantFields, bool) {
	f, err := validateAssistantPayload(rawName, rawDesc, rawModel, rawSystem, rawExample)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return assistantFields{}, false
	}

	return f, true
}

func validateAssistantPayload(rawName, rawDesc, rawModel, rawSystem string, rawExample *string) (assistantFields, error) {
	name, err := httpx.RequireField("name", rawName, maxAssistantNameLen)
	if err != nil {
		return assistantFields{}, err
	}
	description, err := httpx.RequireField("description", rawDesc, maxAssistantDescriptionLen)
	if err != nil {
		return assistantFields{}, err
	}
	model, err := httpx.RequireField("model", rawModel, maxAssistantModelLen)
	if err != nil {
		return assistantFields{}, err
	}
	systemPrompt, err := httpx.RequireField("systemPrompt", rawSystem, maxAssistantSystemPromptLen)
	if err != nil {
		return assistantFields{}, err
	}

	var examplePrompt *string
	if rawExample != nil {
		ex := strings.TrimSpace(*rawExample)
		if len(ex) > maxAssistantExamplePromptLen {
			return assistantFields{}, errors.New("exampleUserPrompt is too long")
		}
		if ex != "" {
			examplePrompt = &ex
		}
	}

	return assistantFields{
		Name: name, Description: description, Model: model,
		SystemPrompt: systemPrompt, ExamplePrompt: examplePrompt,
	}, nil
}

func toAPIAssistant(a domain.Assistant, hideSystemPrompt bool) api.Assistant {
	out := api.Assistant{
		Id:                a.ID,
		CategoryId:        a.CategoryID,
		CategoryName:      a.CategoryName,
		Name:              a.Name,
		Description:       a.Description,
		Model:             a.Model,
		ExampleUserPrompt: a.ExampleUserPrompt,
		IsActive:          a.IsActive,
		CreatedAt:         new(a.CreatedAt.UTC()),
		UpdatedAt:         new(a.UpdatedAt.UTC()),
	}
	if !hideSystemPrompt {
		out.SystemPrompt = new(a.SystemPrompt)
	}

	return out
}
