package auth

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	maxEmailLen       = 254
	minPasswordRunes  = 8
	maxPasswordBytes  = 72
	passwordHashCost  = 12
	dummyHashPassword = "dummy-password-0"
)

var dummyPasswordHash = mustHashPassword(dummyHashPassword)

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", errors.New("email is required")
	}
	if len(email) > maxEmailLen {
		return "", errors.New("email is too long")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", errors.New("email is invalid")
	}

	return email, nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if utf8.RuneCountInString(password) < minPasswordRunes {
		return errors.New("password must contain at least 8 characters")
	}
	if len(password) > maxPasswordBytes {
		return errors.New("password is too long")
	}

	var hasLetter, hasNumberOrSpecial bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasNumberOrSpecial = true
		}
	}
	if !hasLetter {
		return errors.New("password must contain at least one letter")
	}
	if !hasNumberOrSpecial {
		return errors.New("password must contain at least one digit or special character")
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func burnPasswordCheck(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
}

func mustHashPassword(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		panic(err)
	}

	return hash
}
