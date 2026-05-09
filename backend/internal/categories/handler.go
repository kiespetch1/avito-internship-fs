package categories

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/httpx"
)

const maxCategoryNameLen = 256

type Service interface {
	List(ctx context.Context) ([]domain.Category, error)
	Create(ctx context.Context, name string, description *string) (domain.Category, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

type listResponse struct {
	Categories []api.Category `json:"categories"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to list categories")
		return
	}

	resp := listResponse{Categories: make([]api.Category, 0, len(cats))}
	for _, c := range cats {
		resp.Categories = append(resp.Categories, toAPICategory(c))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req api.PostCategoriesJSONRequestBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "malformed request body")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "name is required")
		return
	}
	if len(name) > maxCategoryNameLen {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "name is too long")
		return
	}

	c, err := h.svc.Create(r.Context(), name, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNameTaken) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "category name already exists")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to create category")

		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toAPICategory(c))
}

func toAPICategory(c domain.Category) api.Category {
	created := c.CreatedAt.UTC()

	return api.Category{
		Id:          openapi_types.UUID(c.ID),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   &created,
	}
}
