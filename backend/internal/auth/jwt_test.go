package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParse(t *testing.T) {
	iss, err := NewIssuer("secret", time.Hour)
	if err != nil {
		t.Fatalf("new issuer: %v", err)
	}
	id := uuid.New()
	tok, err := iss.Issue(id, RoleAdmin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := iss.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != id.String() {
		t.Fatalf("subject mismatch: got %q want %q", claims.Subject, id.String())
	}
	if claims.Role != RoleAdmin {
		t.Fatalf("role mismatch: got %q", claims.Role)
	}
}

func TestParseRejectsBadSecret(t *testing.T) {
	iss, _ := NewIssuer("secret-a", time.Hour)
	tok, _ := iss.Issue(uuid.New(), RoleUser)

	other, _ := NewIssuer("secret-b", time.Hour)
	if _, err := other.Parse(tok); err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	iss, _ := NewIssuer("secret", time.Hour)
	iss.ttl = -time.Minute
	tok, _ := iss.Issue(uuid.New(), RoleUser)
	if _, err := iss.Parse(tok); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestNewIssuerRejectsEmptySecret(t *testing.T) {
	if _, err := NewIssuer("", time.Hour); err == nil {
		t.Fatal("expected error for empty secret")
	}
}
