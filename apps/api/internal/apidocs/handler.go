package apidocs

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

type Handler struct {
	path   string
	logger *slog.Logger
}

func NewHandler(path string, logger *slog.Logger) *Handler {
	return &Handler{path: path, logger: logger}
}

func (h *Handler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(h.path)
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Wrap(err))
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(content)
}

func (h *Handler) Docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, `<!doctype html><html><head><title>Rail Booking API</title><meta charset="utf-8"><meta name="viewport" content="width=device-width"></head><body><h1>Segment Train Booking API</h1><p><a href="/openapi.yaml">Open OpenAPI specification</a></p></body></html>`)
}
