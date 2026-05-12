package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPageParamsDefaults(t *testing.T) {
	page, size, err := PageParams(url.Values{}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 1 || size != 10 {
		t.Fatalf("defaults: got page=%d size=%d", page, size)
	}
}

func TestPageParamsParsesValues(t *testing.T) {
	q := url.Values{}
	q.Set("page", "3")
	q.Set("pageSize", "25")
	page, size, err := PageParams(q, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 3 || size != 25 {
		t.Fatalf("parsed: got page=%d size=%d", page, size)
	}
}

func TestPageParamsRejectsNonInteger(t *testing.T) {
	q := url.Values{}
	q.Set("page", "abc")
	_, _, err := PageParams(q, 10)
	if !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("expected ErrInvalidPagination, got %v", err)
	}
}

func TestPageParamsRejectsZero(t *testing.T) {
	q := url.Values{}
	q.Set("pageSize", "0")
	_, _, err := PageParams(q, 10)
	if !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("expected ErrInvalidPagination, got %v", err)
	}
}

func TestPageParamsRejectsNegative(t *testing.T) {
	q := url.Values{}
	q.Set("page", "-1")
	_, _, err := PageParams(q, 10)
	if !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("expected ErrInvalidPagination, got %v", err)
	}
}

func TestRequireFieldTrimsAndReturns(t *testing.T) {
	v, err := RequireField("name", "  hello  ", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hello" {
		t.Fatalf("trim failed: %q", v)
	}
}

func TestRequireFieldRejectsEmpty(t *testing.T) {
	_, err := RequireField("name", "   ", 10)
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("expected ErrInvalidField, got %v", err)
	}
}

func TestRequireFieldRejectsTooLong(t *testing.T) {
	_, err := RequireField("name", "abcdef", 3)
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("expected ErrInvalidField, got %v", err)
	}
}

func TestDecodeJSONSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"name":"x"}`))
	var dst struct {
		Name string `json:"name"`
	}
	if !DecodeJSON(rr, req, &dst) {
		t.Fatalf("decode should succeed, got status=%d", rr.Code)
	}
	if dst.Name != "x" {
		t.Fatalf("decoded value: %q", dst.Name)
	}
}

func TestDecodeJSONMalformed(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`not json`))
	var dst map[string]any
	if DecodeJSON(rr, req, &dst) {
		t.Fatal("decode should fail")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"name":"x","extra":1}`))
	var dst struct {
		Name string `json:"name"`
	}
	if DecodeJSON(rr, req, &dst) {
		t.Fatal("decode should reject unknown fields")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestWriteJSONSetsContentTypeAndStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusAccepted, map[string]string{"hello": "world"})

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: %q", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["hello"] != "world" {
		t.Fatalf("body: %v", got)
	}
}

func TestWriteErrorEncodesCodeAndMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusForbidden, CodeForbidden, "no access")

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: %d", rr.Code)
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != string(CodeForbidden) || body.Error.Message != "no access" {
		t.Fatalf("error body: %+v", body)
	}
}
