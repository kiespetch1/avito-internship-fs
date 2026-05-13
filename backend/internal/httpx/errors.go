package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorCode string

const (
	CodeInvalidRequest     ErrorCode = "INVALID_REQUEST"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeCategoryNotFound   ErrorCode = "CATEGORY_NOT_FOUND"
	CodeAssistantNotFound  ErrorCode = "ASSISTANT_NOT_FOUND"
	CodeAssistantInactive  ErrorCode = "ASSISTANT_INACTIVE"
	CodeLLMProviderError   ErrorCode = "LLM_PROVIDER_ERROR"
	CodeInternalError      ErrorCode = "INTERNAL_ERROR"
)

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorBody{Error: errorPayload{Code: code, Message: message}}); err != nil {
		slog.Error("write error response failed", "error", err)
	}
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write json response failed", "error", err)
	}
}
