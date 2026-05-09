package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/httpx"
)

var (
	DummyAdminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	DummyUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

const (
	dummyAdminEmail = "admin@test.local"
	dummyUserEmail  = "user@test.local"
)

type Handler struct {
	issuer *Issuer
	now    func() time.Time
}

func NewHandler(issuer *Issuer) *Handler {
	return &Handler{issuer: issuer, now: time.Now}
}

func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req api.PostDummyLoginJSONRequestBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "malformed request body")
		return
	}
	role, ok := roleFromAPI(req.Role)
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "role must be 'admin' or 'user'")
		return
	}

	var (
		id    uuid.UUID
		email string
	)
	switch role {
	case RoleAdmin:
		id, email = DummyAdminID, dummyAdminEmail
	case RoleUser:
		id, email = DummyUserID, dummyUserEmail
	}

	token, err := h.issuer.Issue(id, role)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to issue token")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, api.Token{
		Token: token,
		User: api.User{
			Id:    id,
			Email: openapi_types.Email(email),
			Role:  roleToAPI(role),
		},
	})
}
