package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/config"
)

func testRouter() http.Handler {
	return NewRouter(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), config.Config{AllowedOrigins: []string{"http://localhost:3000"}, DatabaseTimeout: time.Second, OpenAPIPath: "../../../../docs/openapi.yaml"})
}
func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	testRouter().ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("got %d", recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected JSON")
	}
}
func TestCORS(t *testing.T) {
	for _, tc := range []struct{ origin, want string }{{"http://localhost:3000", "http://localhost:3000"}, {"https://evil.example", ""}} {
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		request.Header.Set("Origin", tc.origin)
		recorder := httptest.NewRecorder()
		testRouter().ServeHTTP(recorder, request)
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
		if tc.want != "" && recorder.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("expected credentialed CORS for an allowed origin")
		}
	}
}

func TestPassengerBookingsRequiresAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/passenger/bookings", nil)
	recorder := httptest.NewRecorder()
	testRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got %d; want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestSeatMapRouteRejectsInvalidRunID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/train-runs/not-a-uuid/seat-map?originStationId=20000000-0000-4000-8000-000000000001&destinationStationId=20000000-0000-4000-8000-000000000005", nil)
	recorder := httptest.NewRecorder()
	testRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got %d; want %d", recorder.Code, http.StatusBadRequest)
	}
}
