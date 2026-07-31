package trainrun

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
func (h *Handler) Routes(r chi.Router) {
	r.Get("/train-runs", h.list)
	r.Get("/train-runs/{trainRunId}", h.get)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	date, err := time.Parse("2006-01-02", r.URL.Query().Get("travelDate"))
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("travelDate", "A valid travelDate is required."))
		return
	}
	filters := Filters{TravelDate: date}
	if raw := r.URL.Query().Get("routeId"); raw != "" {
		id, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			httpx.WriteError(w, r, h.logger, apperror.Validation("routeId", "Must be a UUID."))
			return
		}
		filters.RouteID = &id
	}
	items, err := h.service.List(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, httpx.PageOf(items))
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "trainRunId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, item)
}
