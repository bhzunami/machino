package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhzunami/machino/internal/handler"
	"github.com/bhzunami/machino/internal/mailer"
	"github.com/bhzunami/machino/internal/realtime"
	"github.com/bhzunami/machino/internal/store"
)

func main() {
	setAdminEmail := flag.String("set-admin", "", "grant admin role to the user with this email, then exit")
	flag.Parse()

	logLevel := slog.LevelInfo
	switch env("LOG_LEVEL", "info") {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	var logHandler slog.Handler
	if env("LOG_FORMAT", "text") == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}
	logger := slog.New(logHandler)
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
	if *setAdminEmail != "" {
		user, err := s.SetAdminByEmail(ctx, *setAdminEmail)
		if err != nil {
			logger.Error("set admin", "email", *setAdminEmail, "error", err)
			os.Exit(1)
		}
		logger.Info("admin role granted", "email", user.Email, "user_id", user.ID)
		return
	}

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

	registrationEnabled := env("REGISTRATION_ENABLED", "true") != "false"
	cookieSecure := env("COOKIE_SECURE", "false") == "true"
	if !registrationEnabled {
		logger.Info("registration disabled — set REGISTRATION_ENABLED=true to enable")
	}
	if cookieSecure {
		logger.Info("secure cookies enabled — HSTS will be sent")
	}

	h := handler.New(s, hub, m, logger).
		WithRegistration(registrationEnabled).
		WithCookieSecure(cookieSecure)
	router := h.Router(staticDir)
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
	h.Shutdown()
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
