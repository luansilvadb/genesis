package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("expected empty default database URL (explicit config required), got: %s", cfg.DatabaseURL)
	}
	if cfg.JWTExpiry.Hours() != 24 {
		t.Errorf("expected JWT expiry 24h, got %v", cfg.JWTExpiry)
	}
	if cfg.SMTPHost != "smtp.zoho.com" {
		t.Errorf("expected SMTP host smtp.zoho.com, got %s", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("expected SMTP port 587, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUseTLS != false {
		t.Errorf("expected SMTP TLS to be false by default (STARTTLS on port 587)")
	}
	if cfg.FrontendURL != "http://localhost:5173" {
		t.Errorf("expected frontend URL http://localhost:5173, got %s", cfg.FrontendURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE_URL", "postgres://custom:custom@db:5432/test")
	os.Setenv("JWT_SECRET", "a-valid-jwt-secret-with-sufficient-length-for-testing")
	os.Setenv("SMTP_HOST", "smtp.gmail.com")
	os.Setenv("SMTP_PORT", "465")
	os.Setenv("SMTP_USER", "user@test.com")
	os.Setenv("SMTP_PASS", "secret123")
	os.Setenv("SMTP_USE_TLS", "false")
	os.Setenv("FRONTEND_URL", "https://app.example.com")
	os.Setenv("GOOGLE_CLIENT_ID", "google-id-123")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USER")
		os.Unsetenv("SMTP_PASS")
		os.Unsetenv("SMTP_USE_TLS")
		os.Unsetenv("FRONTEND_URL")
		os.Unsetenv("GOOGLE_CLIENT_ID")
	}()

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://custom:custom@db:5432/test" {
		t.Errorf("unexpected database URL: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "a-valid-jwt-secret-with-sufficient-length-for-testing" {
		t.Errorf("expected JWT secret, got %s", cfg.JWTSecret)
	}
	if cfg.SMTPHost != "smtp.gmail.com" {
		t.Errorf("expected SMTP host smtp.gmail.com, got %s", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 465 {
		t.Errorf("expected SMTP port 465, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "user@test.com" {
		t.Errorf("expected SMTP user user@test.com, got %s", cfg.SMTPUser)
	}
	if cfg.SMTPPass != "secret123" {
		t.Errorf("expected SMTP pass 'secret123', got %s", cfg.SMTPPass)
	}
	if cfg.SMTPUseTLS != false {
		t.Errorf("expected SMTP TLS to be false when SMTP_USE_TLS=false")
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("expected frontend URL https://app.example.com, got %s", cfg.FrontendURL)
	}
	if cfg.GoogleOAuthID != "google-id-123" {
		t.Errorf("expected Google OAuth ID 'google-id-123', got %s", cfg.GoogleOAuthID)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	cfg := Load()
	if cfg.Port != 0 {
		t.Errorf("expected port 0 for invalid input, got %d", cfg.Port)
	}
}

func TestLoadEmptyEnvDoesNotOverride(t *testing.T) {
	os.Setenv("JWT_SECRET", "")
	os.Setenv("SMTP_USER", "")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SMTP_USER")
	}()

	cfg := Load()
	if cfg.JWTSecret != "" {
		t.Errorf("expected JWT secret to be empty, got %s", cfg.JWTSecret)
	}
	if cfg.SMTPUser != "" {
		t.Errorf("expected SMTP user to be empty, got %s", cfg.SMTPUser)
	}
}
