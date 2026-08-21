package slogchi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFilterEarly_BodyCaptureSkipped(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	// Create a middleware that accepts only /foo but logs /bar
	mw := NewWithConfig(logger, Config{
		Filters:       []Filter{AcceptPathPrefix("/bar")},
		WithRequestBody: true,
	})

	// Create a large body
	largeBody := strings.Repeat("x", 10000)
	req := httptest.NewRequest(http.MethodPost, "/foo", strings.NewReader(largeBody))
	w := httptest.NewRecorder()

	// The handler should still process the request (filter only affects logging)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	// No log should be emitted for /foo
	if logBuf.Len() > 0 {
		t.Errorf("Expected no log output for filtered path, got: %s", logBuf.String())
	}
}

func TestFilterEarly_StatusFilter(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	// Create a middleware that only logs 4xx status codes
	mw := NewWithConfig(logger, Config{
		Filters:       []Filter{AcceptStatusLessThan(500)},
		WithRequestBody: true,
	})

	// Test with 200 OK
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	// Should log the 200 response
	if logBuf.Len() == 0 {
		t.Error("Expected log output for 200 response with AcceptStatusLessThan(500) filter")
	}

	logBuf.Reset()

	// Test with 500 error
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})).ServeHTTP(w, req)

	// Should NOT log the 500 response
	if logBuf.Len() > 0 {
		t.Errorf("Expected no log output for 500 response with AcceptStatusLessThan(500) filter, got: %s", logBuf.String())
	}
}

func TestFilterEarly_NoFilterLogsAll(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	// No filters - should log everything
	mw := NewWithConfig(logger, DefaultConfig())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if logBuf.Len() == 0 {
		t.Error("Expected log output when no filters are configured")
	}
}
