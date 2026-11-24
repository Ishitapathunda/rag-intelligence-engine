package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "healthy") {
		t.Fatalf("expected healthy response, got %s", rec.Body.String())
	}
}

func TestQueryRejectsEmptyQuestion(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/query",
		strings.NewReader(`{"question":"   "}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestQueryRejectsInvalidJSON(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/query",
		strings.NewReader(`invalid json`),
	)

	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestQueryRejectsUnknownFields(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/query",
		strings.NewReader(`{"question":"hello","extra":"value"}`),
	)

	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
