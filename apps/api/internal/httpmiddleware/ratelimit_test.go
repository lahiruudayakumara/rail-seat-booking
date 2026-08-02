package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitRejectsRequestsAboveLimit(t *testing.T) {
	handler := RateLimit(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for attempt, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodPost, "/bookings/access", nil)
		request.RemoteAddr = "192.0.2.10"
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != want {
			t.Fatalf("attempt %d: got %d; want %d", attempt+1, recorder.Code, want)
		}
	}
}

func TestRateLimitSeparatesClients(t *testing.T) {
	handler := RateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, client := range []string{"192.0.2.10", "192.0.2.11"} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = client
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("client %s unexpectedly limited", client)
		}
	}
}
