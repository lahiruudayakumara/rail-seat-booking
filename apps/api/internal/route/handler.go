package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
	"log/slog"
	"net/http"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
func (h *Handler) Routes(r chi.Router) { r.Get("/routes", h.list); r.Get("/routes/{routeId}", h.get) }
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, httpx.PageOf(items))
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "routeId")
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
