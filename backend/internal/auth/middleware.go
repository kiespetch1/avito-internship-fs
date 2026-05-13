package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"avito-internship-fs/internal/httpx"
)

type ctxKey int

const principalKey ctxKey = iota

type Principal struct {
	UserID uuid.UUID
	Role   Role
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

func RequireAuth(issuer *Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "missing authorization header")
				return
			}
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid authorization header")
				return
			}
			claims, err := issuer.Parse(parts[1])
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid or expired token")
				return
			}
			userID, _ := uuid.Parse(claims.UserID)
			ctx := context.WithValue(r.Context(), principalKey, Principal{UserID: userID, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "not authenticated")
				return
			}
			for _, role := range roles {
				if p.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "insufficient role")
		})
	}
}
