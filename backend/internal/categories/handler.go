package categories

import (
	"context"
	"errors"
	"net/http"

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
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	name, err := httpx.RequireField("name", req.Name, maxCategoryNameLen)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
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
	return api.Category{
		Id:          openapi_types.UUID(c.ID),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   new(c.CreatedAt.UTC()),
	}
}
