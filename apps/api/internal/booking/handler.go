package booking

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/httpmiddleware"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
func (h *Handler) Routes(r chi.Router) {
	r.With(httpmiddleware.RateLimit(30, time.Minute)).Post("/bookings", h.create)
	r.With(httpmiddleware.RateLimit(60, time.Minute)).Post("/booking-holds", h.createHold)
	r.With(httpmiddleware.RateLimit(10, time.Minute)).Post("/bookings/access", h.access)
	r.Post("/bookings/{bookingId}/cancel", h.cancel)
}
func (h *Handler) createHold(w http.ResponseWriter, r *http.Request) {
	var request HoldRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	hold, err := h.service.CreateHold(r.Context(), request, httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, hold)
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "Request body is too large."))
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var request CreateRequest
	if err = decoder.Decode(&request); err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "Invalid JSON body."))
		return
	}
	item, replayed, err := h.service.Create(r.Context(), request, strings.TrimSpace(r.Header.Get("Idempotency-Key")), payload, httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
		httpx.WriteJSON(w, 200, item)
		return
	}
	w.Header().Set("Location", "/api/v1/bookings/"+item.ID.String())
	httpx.WriteJSON(w, 201, item)
}
func (h *Handler) access(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	var request AccessRequest
	if err := decoder.Decode(&request); err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "Invalid JSON body."))
		return
	}
	item, err := h.service.Access(r.Context(), request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, item)
}
func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "bookingId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	item, err := h.service.Cancel(r.Context(), id, token, httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, item)
}
