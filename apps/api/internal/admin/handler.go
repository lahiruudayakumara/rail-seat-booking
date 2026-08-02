package admin

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
	apiKey  string
}

func NewHandler(service *Service, logger *slog.Logger, apiKey string) *Handler {
	return &Handler{service: service, logger: logger, apiKey: apiKey}
}

func (h *Handler) Routes(r chi.Router) {
	r.With(h.authorize).Get("/admin/train-runs/{trainRunId}/dashboard", h.dashboard)
}

func (h *Handler) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if len(provided) != len(h.apiKey) || subtle.ConstantTimeCompare([]byte(provided), []byte(h.apiKey)) != 1 {
			httpx.WriteError(w, r, h.logger, apperror.New(401, "ADMIN_ACCESS_DENIED", "Administrator authentication is required.", nil))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "trainRunId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	item, err := h.service.Dashboard(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
