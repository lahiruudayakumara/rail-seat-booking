package passengerauth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	r.With(h.RequireSession).Get("/passenger/travellers", h.listTravellers)
	r.With(h.RequireSession).Post("/passenger/travellers", h.createTraveller)
	r.With(h.RequireSession).Put("/passenger/travellers/{travellerID}", h.updateTraveller)
	r.With(h.RequireSession).Delete("/passenger/travellers/{travellerID}", h.deleteTraveller)
	r.With(h.RequireSession).Get("/passenger/preferences", h.getPreferences)
	r.With(h.RequireSession).Put("/passenger/preferences", h.updatePreferences)
}

func (h *Handler) listTravellers(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	travellers, err := h.service.ListTravellers(r.Context(), account.ID)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": travellers})
}

func (h *Handler) createTraveller(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	var request TravellerRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	traveller, err := h.service.CreateTraveller(r.Context(), account.ID, request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	w.Header().Set("Location", "/api/v1/passenger/travellers/"+traveller.ID.String())
	httpx.WriteJSON(w, http.StatusCreated, traveller)
}

func (h *Handler) updateTraveller(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	travellerID, err := uuid.Parse(chi.URLParam(r, "travellerID"))
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("travellerID", "Traveller ID must be a UUID."))
		return
	}
	var request TravellerRequest
	if decodeErr := httpx.DecodeJSON(w, r, &request); decodeErr != nil {
		httpx.WriteError(w, r, h.logger, decodeErr)
		return
	}
	traveller, err := h.service.UpdateTraveller(r.Context(), account.ID, travellerID, request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, traveller)
}

func (h *Handler) deleteTraveller(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	travellerID, err := uuid.Parse(chi.URLParam(r, "travellerID"))
	if err != nil {
		httpx.WriteError(w, r, h.logger, apperror.Validation("travellerID", "Traveller ID must be a UUID."))
		return
	}
	if err = h.service.DeleteTraveller(r.Context(), account.ID, travellerID); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	preferences, err := h.service.GetPreferences(r.Context(), account.ID)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, preferences)
}

func (h *Handler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	account, _ := AccountFromContext(r.Context())
	var request PreferencesRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	preferences, err := h.service.UpdatePreferences(r.Context(), account.ID, request)
	if err != nil {
		httpx.WriteError(w, r, h.logger, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, preferences)
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
