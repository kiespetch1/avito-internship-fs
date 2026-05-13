package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/domain"
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
	users  UserStore
}

type UserStore interface {
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

func NewHandler(issuer *Issuer, users UserStore) *Handler {
	return &Handler{issuer: issuer, users: users}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "user store is unavailable")
		return
	}

	creds, ok := readCredentials(w, r)
	if !ok {
		return
	}

	passwordHash, err := hashPassword(creds.Password)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to hash password")
		return
	}

	user, err := h.users.Create(r.Context(), creds.Email, passwordHash)
	if err != nil {
		if errors.Is(err, domain.ErrUserEmailTaken) {
			httpx.WriteError(w, http.StatusConflict, httpx.CodeEmailAlreadyExists, "email already registered")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to create user")

		return
	}

	h.writeToken(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "user store is unavailable")
		return
	}

	creds, ok := readCredentials(w, r)
	if !ok {
		return
	}

	user, err := h.users.GetByEmail(r.Context(), creds.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			burnPasswordCheck(creds.Password)
			writeInvalidCredentials(w)

			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to load user")

		return
	}
	if !checkPassword(user.PasswordHash, creds.Password) {
		writeInvalidCredentials(w)
		return
	}

	h.writeToken(w, http.StatusOK, user)
}

func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req api.PostDummyLoginJSONRequestBody
	if !httpx.DecodeJSON(w, r, &req) {
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

type credentials struct {
	Email    string
	Password string
}

func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var req api.AuthCredentialsIn
	if !httpx.DecodeJSON(w, r, &req) {
		return credentials{}, false
	}

	email, err := normalizeEmail(string(req.Email))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return credentials{}, false
	}
	if err := validatePassword(req.Password); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidRequest, err.Error())
		return credentials{}, false
	}

	return credentials{Email: email, Password: req.Password}, true
}

func (h *Handler) writeToken(w http.ResponseWriter, status int, user domain.User) {
	role, ok := roleFromString(user.Role)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "user role is invalid")
		return
	}
	token, err := h.issuer.Issue(user.ID, role)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "failed to issue token")
		return
	}

	createdAt := user.CreatedAt.UTC()
	httpx.WriteJSON(w, status, api.Token{
		Token: token,
		User: api.User{
			Id:        user.ID,
			Email:     openapi_types.Email(user.Email),
			Role:      roleToAPI(role),
			CreatedAt: &createdAt,
		},
	})
}

func writeInvalidCredentials(w http.ResponseWriter) {
	httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid email or password")
}

func roleFromString(raw string) (Role, bool) {
	switch Role(raw) {
	case RoleAdmin:
		return RoleAdmin, true
	case RoleUser:
		return RoleUser, true
	default:
		return "", false
	}
}
