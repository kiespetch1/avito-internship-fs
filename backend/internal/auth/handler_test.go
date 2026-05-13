package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/domain"
)

func newTestHandler(t *testing.T) (*Handler, *Issuer, *fakeUserStore) {
	t.Helper()
	iss, err := NewIssuer("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("issuer: %v", err)
	}
	users := newFakeUserStore()

	return NewHandler(iss, users), iss, users
}

func TestDummyLoginAdmin(t *testing.T) {
	h, iss, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader(`{"role":"admin"}`))
	rr := httptest.NewRecorder()
	h.DummyLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp api.Token
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.User.Id != DummyAdminID {
		t.Fatalf("admin id: got %q", resp.User.Id)
	}
	if resp.User.Role != api.RoleAdmin {
		t.Fatalf("role: got %q", resp.User.Role)
	}
	claims, err := iss.Parse(resp.Token)
	if err != nil {
		t.Fatalf("parse issued token: %v", err)
	}
	if claims.Subject != DummyAdminID.String() || claims.Role != RoleAdmin {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestDummyLoginUser(t *testing.T) {
	h, _, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader(`{"role":"user"}`))
	rr := httptest.NewRecorder()
	h.DummyLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var resp api.Token
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.User.Id != DummyUserID {
		t.Fatalf("user id: got %q", resp.User.Id)
	}
}

func TestDummyLoginInvalidRole(t *testing.T) {
	h, _, _ := newTestHandler(t)
	cases := []string{
		`{"role":"superadmin"}`,
		`{}`,
		`not json`,
		`{"role":"admin","extra":1}`,
	}
	for _, body := range cases {
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", strings.NewReader(body))
		rr := httptest.NewRecorder()
		h.DummyLogin(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("body=%q: got status %d, want 400", body, rr.Code)
		}
	}
}

func TestRegisterCreatesUserToken(t *testing.T) {
	h, iss, store := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":" New.User@Example.COM ","password":"passw0rd"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d want 201, body=%s", rr.Code, rr.Body.String())
	}
	var resp api.Token
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.User.Email != "new.user@example.com" {
		t.Fatalf("email normalized: got %q", resp.User.Email)
	}
	if resp.User.Role != api.RoleUser {
		t.Fatalf("registered role: got %q", resp.User.Role)
	}
	stored := store.users["new.user@example.com"]
	if stored.PasswordHash == "" || stored.PasswordHash == "passw0rd" {
		t.Fatalf("password must be stored as hash, got %q", stored.PasswordHash)
	}
	if !checkPassword(stored.PasswordHash, "passw0rd") {
		t.Fatalf("stored hash does not verify original password")
	}
	claims, err := iss.Parse(resp.Token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Subject != resp.User.Id.String() || claims.Role != RoleUser {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestRegisterRejectsRoleField(t *testing.T) {
	h, _, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"u@example.com","password":"passw0rd","role":"admin"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	h, _, _ := newTestHandler(t)
	body := `{"email":"u@example.com","password":"passw0rd"}`

	rr := httptest.NewRecorder()
	h.Register(rr, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("first register: status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.Register(rr, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body)))
	if rr.Code != http.StatusConflict {
		t.Fatalf("second register: got %d want 409 body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	h, _, _ := newTestHandler(t)
	cases := []string{
		`{"email":"u@example.com","password":"short1"}`,
		`{"email":"u@example.com","password":"password"}`,
		`{"email":"u@example.com","password":"12345678"}`,
	}
	for _, body := range cases {
		rr := httptest.NewRecorder()
		h.Register(rr, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body)))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("body=%s: got %d want 400", body, rr.Code)
		}
	}
}

func TestLoginWithEmailPassword(t *testing.T) {
	h, iss, store := newTestHandler(t)
	hash, err := hashPassword("passw0rd")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	userID := uuid.New()
	store.users["u@example.com"] = domain.User{
		ID: userID, Email: "u@example.com", PasswordHash: hash,
		Role: string(RoleUser), CreatedAt: time.Now(),
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"u@example.com","password":"passw0rd"}`))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp api.Token
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.User.Id != userID {
		t.Fatalf("user id: got %s want %s", resp.User.Id, userID)
	}
	claims, err := iss.Parse(resp.Token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Role != RoleUser || claims.Subject != userID.String() {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	h, _, store := newTestHandler(t)
	hash, err := hashPassword("passw0rd")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	store.users["u@example.com"] = domain.User{
		ID: uuid.New(), Email: "u@example.com", PasswordHash: hash,
		Role: string(RoleUser), CreatedAt: time.Now(),
	}

	cases := []string{
		`{"email":"u@example.com","password":"wrongpass1"}`,
		`{"email":"missing@example.com","password":"wrongpass1"}`,
	}
	for _, body := range cases {
		rr := httptest.NewRecorder()
		h.Login(rr, httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body)))
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("body=%s: got %d want 401 body=%s", body, rr.Code, rr.Body.String())
		}
	}
}

type fakeUserStore struct {
	users map[string]domain.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{users: make(map[string]domain.User)}
}

func (s *fakeUserStore) Create(_ context.Context, email, passwordHash string) (domain.User, error) {
	if _, exists := s.users[email]; exists {
		return domain.User{}, domain.ErrUserEmailTaken
	}
	u := domain.User{
		ID: uuid.New(), Email: email, PasswordHash: passwordHash,
		Role: string(RoleUser), CreatedAt: time.Now(),
	}
	s.users[email] = u

	return u, nil
}

func (s *fakeUserStore) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := s.users[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	if u.ID == uuid.Nil {
		return domain.User{}, errors.New("test user id is empty")
	}

	return u, nil
}
