package store

import (
	"context"
	"log/slog"
	"testing"

	"github.com/bhzunami/machino/internal/model"
)

func TestAppSettingsBootstrapAndUpdate(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	defaults := model.AppSettings{
		AppDomain:           "machino.example.com",
		RegistrationEnabled: false,
		SMTPHost:            "smtp.example.com",
		SMTPPort:            "2525",
		SMTPUsername:        "mailer",
		SMTPPassword:        "secret",
		SMTPFrom:            "noreply@example.com",
	}
	if err := s.BootstrapAppSettings(ctx, defaults); err != nil {
		t.Fatalf("bootstrap settings: %v", err)
	}
	settings, err := s.AppSettings(ctx)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.AppDomain != defaults.AppDomain || settings.RegistrationEnabled || !settings.SMTPPasswordSet {
		t.Fatalf("unexpected bootstrapped settings: %#v", settings)
	}

	if _, err := s.UpdateAppSettings(ctx, model.AppSettings{
		AppDomain:           "new.example.com",
		RegistrationEnabled: true,
		SMTPHost:            "smtp2.example.com",
		SMTPPort:            "587",
		SMTPUsername:        "mailer2",
		SMTPFrom:            "hello@example.com",
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	updated, err := s.AppSettings(ctx)
	if err != nil {
		t.Fatalf("load updated settings: %v", err)
	}
	if updated.AppDomain != "new.example.com" || !updated.RegistrationEnabled {
		t.Fatalf("unexpected updated settings: %#v", updated)
	}
	if updated.SMTPPassword != "secret" || !updated.SMTPPasswordSet {
		t.Fatalf("SMTP password should be preserved: %#v", updated)
	}

	if _, err := s.UpdateAppSettings(ctx, model.AppSettings{SMTPPort: "not-a-port"}); err != ErrInvalidInput {
		t.Fatalf("expected invalid input for bad port, got %v", err)
	}
}

func TestCreateSetupAdmin(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	if err := s.BootstrapAppSettings(ctx, model.AppSettings{RegistrationEnabled: true, SMTPPort: "587"}); err != nil {
		t.Fatalf("bootstrap settings: %v", err)
	}

	required, err := s.SetupRequired(ctx)
	if err != nil {
		t.Fatalf("setup required: %v", err)
	}
	if !required {
		t.Fatal("setup should be required before admin exists")
	}

	admin, err := s.CreateSetupAdmin(ctx, "ADMIN@example.com", "Admin", "hash", model.AppSettings{
		AppDomain:           "machino.example.com",
		RegistrationEnabled: false,
		SMTPPort:            "587",
	})
	if err != nil {
		t.Fatalf("create setup admin: %v", err)
	}
	if admin.Email != "admin@example.com" || admin.Role != model.RoleAdmin {
		t.Fatalf("unexpected setup admin: %#v", admin)
	}
	required, err = s.SetupRequired(ctx)
	if err != nil {
		t.Fatalf("setup required after admin: %v", err)
	}
	if required {
		t.Fatal("setup should not be required after admin exists")
	}
	if _, err := s.CreateSetupAdmin(ctx, "second@example.com", "Second", "hash", model.AppSettings{SMTPPort: "587"}); err != ErrUnauthorized {
		t.Fatalf("expected setup to be blocked after admin exists, got %v", err)
	}
}
