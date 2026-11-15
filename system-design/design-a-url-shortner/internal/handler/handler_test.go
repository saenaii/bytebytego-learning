package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/metrics"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/repository"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/service"
)

type handlerStore struct {
	record repository.Record
	found  bool
}

func (s *handlerStore) FindByKey(context.Context, string, string) (repository.Record, bool, error) {
	return repository.Record{}, false, nil
}
func (s *handlerStore) CreateOrGet(_ context.Context, record repository.Record) (repository.Record, bool, error) {
	record.ID, record.ShortCode = 1, "1"
	s.record = record
	return record, false, nil
}
func (s *handlerStore) FindByCode(context.Context, string) (repository.Record, bool, error) {
	return s.record, s.found, nil
}
func (s *handlerStore) Ping(context.Context) error { return nil }

type handlerCache struct{}

func (handlerCache) Get(context.Context, string) (string, bool, error) { return "", false, nil }
func (handlerCache) Set(context.Context, string, string) error         { return nil }
func (handlerCache) Ping(context.Context) error                        { return nil }

func newTestHandler(found bool) http.Handler {
	store := &handlerStore{found: found, record: repository.Record{ShortCode: "1", OriginalURL: "https://example.com"}}
	svc := service.New(store, handlerCache{}, &metrics.Metrics{}, "http://localhost:8080", "full", nil)
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("dashboard")},
		"styles.css": &fstest.MapFile{Data: []byte("body{}")},
		"app.js":     &fstest.MapFile{Data: []byte("// app")},
	}
	return New(svc, assets)
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		found      bool
		wantStatus int
		wantHeader string
	}{
		{name: "dashboard", method: http.MethodGet, path: "/", wantStatus: http.StatusOK},
		{name: "stylesheet", method: http.MethodGet, path: "/styles.css", wantStatus: http.StatusOK},
		{name: "script", method: http.MethodGet, path: "/app.js", wantStatus: http.StatusOK},
		{name: "unknown asset", method: http.MethodGet, path: "/missing.js", wantStatus: http.StatusNotFound},
		{name: "post to asset", method: http.MethodPost, path: "/app.js", wantStatus: http.StatusNotFound},
		{name: "path traversal", method: http.MethodGet, path: "/../go.mod", wantStatus: http.StatusNotFound},
		{name: "shorten", method: http.MethodPost, path: "/api/v1/shorten", body: `{"url":"https://example.com"}`, wantStatus: http.StatusOK},
		{name: "invalid JSON", method: http.MethodPost, path: "/api/v1/shorten", body: "{", wantStatus: http.StatusBadRequest},
		{name: "resolve missing", method: http.MethodGet, path: "/api/v1/resolve/missing", wantStatus: http.StatusNotFound},
		{name: "redirect", method: http.MethodGet, path: "/shorturl/1", found: true, wantStatus: http.StatusFound, wantHeader: "https://example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			newTestHandler(tt.found).ServeHTTP(response, req)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantHeader != "" && response.Header().Get("Location") != tt.wantHeader {
				t.Fatalf("Location = %q, want %q", response.Header().Get("Location"), tt.wantHeader)
			}
		})
	}
}
