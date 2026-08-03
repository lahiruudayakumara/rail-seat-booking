package waitlist

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/httpmiddleware"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/passengerauth"
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
	r.With(httpmiddleware.RateLimit(10, time.Minute)).Post("/waitlist-entries", h.create)
	r.With(httpmiddleware.RateLimit(10, time.Minute)).Post("/waitlist-entries/access", h.access)
	r.Post("/waitlist-entries/{entryId}/cancel", h.cancel)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	entry, err := h.service.Create(r.Context(), request, passengerauth.AccountID(r.Context()))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	w.Header().Set("Location", "/api/v1/waitlist-entries/"+entry.ID.String())
	httpx.WriteJSON(w, http.StatusCreated, entry)
}

func (h *Handler) access(w http.ResponseWriter, r *http.Request) {
	var request AccessRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	entry, err := h.service.Access(r.Context(), request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, entry)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "entryId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	entry, err := h.service.Cancel(r.Context(), id, token)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, entry)
}
