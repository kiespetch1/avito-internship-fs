package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"avito-internship-fs/internal/api"
)

func newTestHandler(t *testing.T) (*Handler, *Issuer) {
	t.Helper()
	iss, err := NewIssuer("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("issuer: %v", err)
	}

	return NewHandler(iss), iss
}

func TestDummyLoginAdmin(t *testing.T) {
	h, iss := newTestHandler(t)
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
	h, _ := newTestHandler(t)
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
	h, _ := newTestHandler(t)
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
