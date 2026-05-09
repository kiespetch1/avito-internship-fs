package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func TestRequireAuthMissingHeader(t *testing.T) {
	iss, _ := NewIssuer("s", time.Hour)
	h := RequireAuth(iss)(okHandler())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rr.Code)
	}
}

func TestRequireAuthInvalidScheme(t *testing.T) {
	iss, _ := NewIssuer("s", time.Hour)
	h := RequireAuth(iss)(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Basic abc")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rr.Code)
	}
}

func TestRequireAuthValid(t *testing.T) {
	iss, _ := NewIssuer("s", time.Hour)
	id := uuid.New()
	tok, _ := iss.Issue(id, RoleUser)

	var seen Principal
	h := RequireAuth(iss)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFrom(r.Context())
		seen = p
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	if seen.UserID != id || seen.Role != RoleUser {
		t.Fatalf("principal mismatch: %+v", seen)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	iss, _ := NewIssuer("s", time.Hour)
	tok, _ := iss.Issue(uuid.New(), RoleUser)

	chain := RequireAuth(iss)(RequireRole(RoleAdmin)(okHandler()))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("got %d want 403", rr.Code)
	}
}

func TestRequireRoleAllowed(t *testing.T) {
	iss, _ := NewIssuer("s", time.Hour)
	tok, _ := iss.Issue(uuid.New(), RoleAdmin)

	chain := RequireAuth(iss)(RequireRole(RoleAdmin)(okHandler()))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("got %d want 204", rr.Code)
	}
}
