package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleUser
}

type Claims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) (*Issuer, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is empty")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return &Issuer{secret: []byte(secret), ttl: ttl}, nil
}

func (i *Issuer) Issue(userID uuid.UUID, role Role) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString(i.secret)
}

func (i *Issuer) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return i.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Role.Valid() {
		return nil, fmt.Errorf("invalid role in token: %q", claims.Role)
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("invalid subject in token: %w", err)
	}

	return claims, nil
}

func roleFromAPI(r api.Role) (Role, bool) {
	switch r {
	case api.RoleAdmin:
		return RoleAdmin, true
	case api.RoleUser:
		return RoleUser, true
	}

	return "", false
}

func roleToAPI(r Role) api.Role {
	switch r {
	case RoleAdmin:
		return api.RoleAdmin
	case RoleUser:
		return api.RoleUser
	}

	return ""
}
