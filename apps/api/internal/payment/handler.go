package payment

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Routes(r chi.Router) { r.Post("/payments/sandbox", h.checkout) }

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	var request CheckoutRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.Checkout(r.Context(), request, strings.TrimSpace(r.Header.Get("Idempotency-Key")), httpx.RequestID(r))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}
