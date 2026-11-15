package handler

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/service"
)

type Handler struct {
	service *service.Service
	assets  fs.FS
}

func New(svc *service.Service, assets fs.FS) *Handler {
	return &Handler{service: svc, assets: assets}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.service.IncMetric("requests")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		h.health(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/metrics":
		h.metrics(w)
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/shorten":
		h.shorten(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/resolve/"):
		h.resolve(w, r, strings.TrimPrefix(r.URL.Path, "/api/v1/resolve/"))
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/shorturl/"):
		h.redirect(w, r, strings.TrimPrefix(r.URL.Path, "/shorturl/"))
	case r.Method == http.MethodGet && r.URL.Path == "/":
		h.dashboard(w, r)
	default:
		h.static(w, r)
	}
}

// static serves embedded dashboard assets that do not match an API route.
func (h *Handler) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if r.Method != http.MethodGet || h.assets == nil || name == "" || name == "." {
		h.writeError(w, http.StatusNotFound, "route not found")
		return
	}
	info, err := fs.Stat(h.assets, name)
	if err != nil || info.IsDir() {
		h.writeError(w, http.StatusNotFound, "route not found")
		return
	}
	http.FileServer(http.FS(h.assets)).ServeHTTP(w, r)
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var request service.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	response, err := h.service.Create(r.Context(), request)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request, code string) {
	if code == "" {
		h.writeError(w, http.StatusNotFound, "short code is required")
		return
	}
	record, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "short code not found")
		return
	}
	h.writeJSON(w, http.StatusOK, record)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request, code string) {
	if code == "" {
		h.writeError(w, http.StatusNotFound, "short code is required")
		return
	}
	record, err := h.service.Resolve(r.Context(), code)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "short code not found")
		return
	}
	h.service.IncMetric("redirect")
	http.Redirect(w, r, record.OriginalURL, http.StatusFound)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.service.Health(ctx); err != nil {
		h.writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unhealthy", "error": err.Error()})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) metrics(w http.ResponseWriter) {
	h.writeJSON(w, http.StatusOK, h.service.Metrics())
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	if h.assets == nil {
		h.writeError(w, http.StatusInternalServerError, "dashboard assets unavailable")
		return
	}
	http.FileServer(http.FS(h.assets)).ServeHTTP(w, r)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
