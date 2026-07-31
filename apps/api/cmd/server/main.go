package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/app"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("parse database configuration", "error", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = cfg.MaxDatabaseConns
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	server := &http.Server{
		Addr:              cfg.APIAddress,
		Handler:           app.NewRouter(pool, logger, cfg),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		logger.Info("api listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("serve", "error", err)
			os.Exit(1)
		}
	}()

	stop, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-stop.Done()
	ctx, release := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer release()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown", "error", err)
	}
}
