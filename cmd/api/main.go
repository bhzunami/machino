package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"machino/internal/handler"
	"machino/internal/mailer"
	"machino/internal/realtime"
	"machino/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPath := env("DATABASE_PATH", "machino.db")
	addr := env("HTTP_ADDR", ":8080")
	staticDir := env("STATIC_DIR", "web/dist")

	s, err := store.Open(ctx, dbPath)
	if err != nil {
		logger.Error("open store", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := s.Close(); err != nil {
			logger.Warn("close store", "error", err)
		}
	}()

	hub := realtime.NewHub(logger)
	m := mailer.New(mailer.Config{
		Host:     env("SMTP_HOST", ""),
		Port:     env("SMTP_PORT", "587"),
		Username: env("SMTP_USERNAME", ""),
		Password: env("SMTP_PASSWORD", ""),
		From:     env("SMTP_FROM", ""),
	})
	if m.Enabled() {
		logger.Info("mailer ready", "host", env("SMTP_HOST", ""), "port", env("SMTP_PORT", "587"))
	} else {
		logger.Warn("mailer not configured — password reset runs in demo mode (token in API response)")
	}
	router := handler.New(s, hub, m, logger).Router(staticDir)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve api", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown api", "error", err)
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
