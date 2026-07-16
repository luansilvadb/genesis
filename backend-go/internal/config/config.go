package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration // Configurable via JWT_EXPIRY_HOURS env var (default 24h)

	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	SMTPUseTLS bool

	FrontendURL   string
	GoogleOAuthID string
	CORSOrigins   []string

	TLSCertFile   string
	TLSKeyFile    string
	EnableSwagger bool
}

func Load() *Config {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	port, _ := strconv.Atoi(getEnv("PORT", "3000"))

	jwtSecret := os.Getenv("JWT_SECRET")
	corsRaw := os.Getenv("CORS_ORIGINS")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	if corsRaw == "" {
		corsRaw = frontendURL
	}

	if isPlaceholderSecret(jwtSecret) {
		log.Fatal("JWT_SECRET contém um valor placeholder. Defina uma chave secreta forte (mínimo 32 caracteres aleatórios).")
	}

	return &Config{
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   jwtSecret,
		JWTExpiry:   time.Duration(getEnvDuration("JWT_EXPIRY_HOURS", 24)) * time.Hour,

		SMTPHost:   getEnv("SMTP_HOST", "smtp.zoho.com"),
		SMTPPort:   smtpPort,
		SMTPUser:   getEnv("SMTP_USER", ""),
		SMTPPass:   getEnv("SMTP_PASS", ""),
		SMTPUseTLS: getEnv("SMTP_USE_TLS", "false") == "true",

		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoogleOAuthID: getEnv("GOOGLE_CLIENT_ID", ""),
		CORSOrigins:   splitEnv(corsRaw),

		TLSCertFile:   getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:    getEnv("TLS_KEY_FILE", ""),
		EnableSwagger: getEnv("SWAGGER_ENABLED", "false") == "true",
	}
}

// isPlaceholderSecret checks whether a JWT secret appears to be a default or
// placeholder value. Secrets matching any of these patterns will cause the
// application to refuse startup, preventing accidental use of weak keys in
// production.
func isPlaceholderSecret(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	placeholders := []string{
		"insira-uma-chave-secreta-aqui",
		"insira-sua-chave-secreta",
		"your-secret-here",
		"your-secret-key",
		"change-me",
		"change-me-please",
		"changeme",
		"replace-me",
		"jwt-secret",
		"jwt_secret",
		"secret",
		"my-secret",
		"super-secret",
	}
	for _, p := range placeholders {
		if lower == p {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallbackHours int) int {
	if v := os.Getenv(key); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			return h
		}
	}
	return fallbackHours
}

func splitEnv(raw string) []string {
	if raw == "" {
		// Safe default: only allow localhost origins in development.
		// Production deployments MUST set CORS_ORIGINS explicitly.
		return []string{"http://localhost:5173"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:5173"}
	}
	return out
}
