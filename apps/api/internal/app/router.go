// Package app is the composition root for the API modular monolith.
package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/apidocs"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/availability"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/config"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/fare"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/health"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/httpmiddleware"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/payment"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/route"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/station"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/trainrun"
)

func NewRouter(pool *pgxpool.Pool, logger *slog.Logger, cfg config.Config) http.Handler {
	journeyService := journey.NewService(journey.NewRepository(pool))
	routeHandler := route.NewHandler(route.NewService(route.NewRepository(pool)), logger)
	stationHandler := station.NewHandler(station.NewService(station.NewRepository(pool)), logger)
	trainRunHandler := trainrun.NewHandler(trainrun.NewService(trainrun.NewRepository(pool)), logger)
	availabilityHandler := availability.NewHandler(availability.NewService(availability.NewRepository(pool), journeyService), logger)
	fareHandler := fare.NewHandler(fare.NewService(fare.NewRepository(pool), journeyService, 5*time.Minute), logger)
	accessSigner := booking.NewAccessSigner(cfg.ManagementSecret, cfg.ManagementTTL)
	bookingRepository := booking.NewRepository()
	paymentProvider := payment.SandboxProvider{}
	bookingHandler := booking.NewHandler(booking.NewService(pool, bookingRepository, accessSigner, payment.NewCancellationProcessor(paymentProvider), cfg.SeatHoldTTL), logger)
	paymentHandler := payment.NewHandler(payment.NewService(pool, payment.NewRepository(), bookingRepository, accessSigner, payment.NewTicketSigner(cfg.ManagementSecret), paymentProvider), logger)
	healthHandler := health.NewHandler(pool, logger, cfg.DatabaseTimeout)
	docsHandler := apidocs.NewHandler(cfg.OpenAPIPath, logger)

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID, chimiddleware.RealIP, chimiddleware.Recoverer)
	router.Use(httpmiddleware.SecureHeaders, httpmiddleware.CORS(cfg.AllowedOrigins), httpmiddleware.AccessLog(logger))
	router.Get("/health", healthHandler.Health)
	router.Get("/ready", healthHandler.Ready)
	router.Get("/openapi.yaml", docsHandler.OpenAPI)
	router.Get("/docs", docsHandler.Docs)
	router.Route("/api/v1", func(r chi.Router) {
		routeHandler.Routes(r)
		stationHandler.Routes(r)
		trainRunHandler.Routes(r)
		availabilityHandler.Routes(r)
		fareHandler.Routes(r)
		bookingHandler.Routes(r)
		paymentHandler.Routes(r)
	})
	return router
}
