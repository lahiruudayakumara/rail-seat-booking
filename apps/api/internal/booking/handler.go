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
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/passengerauth"
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
	r.With(httpmiddleware.RateLimit(15, time.Minute)).Post("/booking-groups", h.createGroup)
	r.With(httpmiddleware.RateLimit(60, time.Minute)).Post("/booking-holds", h.createHold)
	r.Delete("/booking-holds/{holdId}", h.releaseHold)
	r.With(httpmiddleware.RateLimit(10, time.Minute)).Post("/bookings/access", h.access)
	r.Post("/bookings/{bookingId}/cancel", h.cancel)
	r.Get("/passenger/bookings", h.listMine)
}

func (h *Handler) releaseHold(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "holdId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if err = h.service.ReleaseHold(r.Context(), id, token, httpx.RequestID(r)); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "Request body is too large."))
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var request CreateGroupRequest
	if err = decoder.Decode(&request); err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "Invalid JSON body."))
		return
	}
	group, replayed, err := h.service.CreateGroup(r.Context(), request, strings.TrimSpace(r.Header.Get("Idempotency-Key")), payload, httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
		httpx.WriteJSON(w, http.StatusOK, group)
		return
	}
	w.Header().Set("Location", "/api/v1/booking-groups/"+group.ID.String())
	httpx.WriteJSON(w, http.StatusCreated, group)
}
func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) {
	account, ok := passengerauth.AccountFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, h.logger, apperror.New(401, "PASSENGER_AUTH_REQUIRED", "Passenger sign-in is required.", nil))
		return
	}
	items, err := h.service.ListForAccount(r.Context(), account.ID)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
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
	var request CancelRequest
	if r.ContentLength != 0 {
		if err = httpx.DecodeJSON(w, r, &request); err != nil {
			httpx.WriteError(w, r, h.logger, err)
			return
		}
	}
	if len(request.Reason) > 500 {
		httpx.WriteError(w, r, h.logger, apperror.Validation("reason", "Reason must not exceed 500 characters."))
		return
	}
	item, err := h.service.Cancel(r.Context(), id, token, request.Reason, httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, 200, item)
}
