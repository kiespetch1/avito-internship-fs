package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// DecodeJSON читает JSON-тело запроса в dst с DisallowUnknownFields.
// При ошибке сам пишет 400 INVALID_REQUEST и возвращает false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, CodeInvalidRequest, "malformed request body")
		return false
	}

	return true
}

var ErrInvalidField = errors.New("invalid field")

// RequireField возвращает значение после trim, либо ErrInvalidField, если поле пустое или длиннее maxLen
func RequireField(name, raw string, maxLen int) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", fmt.Errorf("%w: %s is required", ErrInvalidField, name)
	}
	if len(v) > maxLen {
		return "", fmt.Errorf("%w: %s is too long", ErrInvalidField, name)
	}

	return v, nil
}
