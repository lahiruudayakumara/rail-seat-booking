package availability

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
func (h *Handler) Routes(r chi.Router) {
	r.Get("/train-runs/{trainRunId}/available-seats", h.list)
	r.Get("/train-runs/{trainRunId}/seat-map", h.seatMap)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	runID, err := httpx.PathUUID(r, "trainRunId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	originID, err := httpx.QueryUUID(r, "originStationId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	destinationID, err := httpx.QueryUUID(r, "destinationStationId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	items, err := h.service.List(r.Context(), runID, originID, destinationID, r.URL.Query().Get("coachClass"))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"trainRunId": runID, "originStationId": originID, "destinationStationId": destinationID, "items": items})
}

func (h *Handler) seatMap(w http.ResponseWriter, r *http.Request) {
	runID, err := httpx.PathUUID(r, "trainRunId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	originID, err := httpx.QueryUUID(r, "originStationId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	destinationID, err := httpx.QueryUUID(r, "destinationStationId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	items, err := h.service.SeatMap(r.Context(), runID, originID, destinationID, r.URL.Query().Get("coachClass"))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"trainRunId": runID, "originStationId": originID, "destinationStationId": destinationID, "items": items})
}
