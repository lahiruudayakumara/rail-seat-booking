package health

import (
	"context"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
	"log/slog"
	"net/http"
	"time"
)

type Pinger interface{ Ping(context.Context) error }
type Handler struct {
	db      Pinger
	logger  *slog.Logger
	timeout time.Duration
}

func NewHandler(db Pinger, logger *slog.Logger, timeout time.Duration) *Handler {
	return &Handler{db: db, logger: logger, timeout: timeout}
}
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, 200, map[string]any{"status": "ok"})
}
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()
	if h.db == nil {
		httpx.WriteError(w, r, h.logger, apperror.New(503, "SERVICE_UNAVAILABLE", "Database is not ready.", nil))
		return
	}
	if err := h.db.Ping(ctx); err != nil {
		httpx.WriteError(w, r, h.logger, apperror.New(503, "SERVICE_UNAVAILABLE", "Database is not ready.", nil))
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"status": "ok", "checks": map[string]string{"database": "ok"}})
}
