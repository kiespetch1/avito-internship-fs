package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

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

type dummyLoginRequest struct {
	Role Role `json:"role"`
}

type userDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Role      Role    `json:"role"`
	CreatedAt *string `json:"createdAt,omitempty"`
}

type tokenResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type Handler struct {
	issuer *Issuer
	now    func() time.Time
}

func NewHandler(issuer *Issuer) *Handler {
	return &Handler{issuer: issuer, now: time.Now}
}

func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "malformed request body")
		return
	}
	if !req.Role.Valid() {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, "role must be 'admin' or 'user'")
		return
	}

	var (
		id    uuid.UUID
		email string
	)
	switch req.Role {
	case RoleAdmin:
		id, email = DummyAdminID, dummyAdminEmail
	case RoleUser:
		id, email = DummyUserID, dummyUserEmail
	}

	token, err := h.issuer.Issue(id, req.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to issue token")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tokenResponse{
		Token: token,
		User: userDTO{
			ID:    id.String(),
			Email: email,
			Role:  req.Role,
		},
	})
}
