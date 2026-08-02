package payment

import (
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
	r.Post("/payments/sandbox", h.checkout)
	r.With(httpmiddleware.RateLimit(20, time.Minute)).Post("/payments/payhere", h.startPayHere)
	r.Get("/payments/{paymentId}", h.payHereStatus)
	r.With(httpmiddleware.RateLimit(240, time.Minute)).Post("/webhooks/payhere", h.payHereWebhook)
	r.With(httpmiddleware.RateLimit(120, time.Minute)).Post("/tickets/verify", h.verifyTicket)
}

func (h *Handler) startPayHere(w http.ResponseWriter, r *http.Request) {
	var request PayHereCheckoutRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.StartPayHereCheckout(r.Context(), request, strings.TrimSpace(r.Header.Get("Idempotency-Key")))
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) payHereStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "paymentId")
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	result, err := h.service.PayHereStatus(r.Context(), id, token)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) payHereWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("body", "A valid PayHere notification form is required."))
		return
	}
	if err := h.service.HandlePayHereWebhook(r.Context(), r.PostForm, httpx.RequestID(r)); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) verifyTicket(w http.ResponseWriter, r *http.Request) {
	var request VerifyTicketRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.VerifyTicket(r.Context(), request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

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
