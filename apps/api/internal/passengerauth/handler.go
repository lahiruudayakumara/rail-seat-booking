package passengerauth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/httpmiddleware"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

const sessionCookie = "rail_passenger_session"

type Handler struct {
	service      *Service
	logger       *slog.Logger
	secureCookie bool
}

func NewHandler(service *Service, logger *slog.Logger, secureCookie bool) *Handler {
	return &Handler{service: service, logger: logger, secureCookie: secureCookie}
}

func (h *Handler) Routes(r chi.Router) {
	r.With(httpmiddleware.RateLimit(5, time.Minute)).Post("/passenger/register", h.register)
	r.With(httpmiddleware.RateLimit(10, time.Minute)).Post("/passenger/login", h.login)
	r.Post("/passenger/logout", h.logout)
	r.With(h.RequireSession).Get("/passenger/me", h.me)
}

func (h *Handler) OptionalSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err == nil {
			account, authErr := h.service.Authenticate(r.Context(), cookie.Value)
			if authErr == nil {
				r = r.WithContext(WithAccount(r.Context(), account))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := AccountFromContext(r.Context()); !ok {
			httpx.WriteError(w, r, h.logger, apperror.New(401, "PASSENGER_AUTH_REQUIRED", "Passenger sign-in is required.", nil))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	account, session, err := h.service.Register(r.Context(), request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	h.setSessionCookie(w, session)
	httpx.WriteJSON(w, http.StatusCreated, account)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	account, session, err := h.service.Login(r.Context(), request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	h.setSessionCookie(w, session)
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err = h.service.Logout(r.Context(), cookie.Value); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, r, h.logger, err)
			return
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, session Session) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: session.Token, Path: "/api/v1", HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode, Expires: session.ExpiresAt, MaxAge: int(time.Until(session.ExpiresAt).Seconds())})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/api/v1", HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}
