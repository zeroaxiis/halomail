// Package config loads service configuration from the environment (and an
// optional .env file). The Config struct is a superset shared by every
// service; each service reads only the sections it needs.
package config

import (
	"fmt"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/usage"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App       App
	HTTP      HTTP
	Postgres  Postgres
	Redis     Redis
	Auth      Auth
	Email     Email
	Google    OAuth `envPrefix:"GOOGLE_"`
	Microsoft Microsoft
	OTel      OTel
	Rate      Rate
	Limits    usage.Policy
}

type App struct {
	Env          string `env:"APP_ENV" envDefault:"development"`
	Name         string `env:"APP_NAME" envDefault:"halomail"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat    string `env:"LOG_FORMAT" envDefault:"json"`
	PublicAPIURL string `env:"PUBLIC_API_URL" envDefault:"http://localhost:8080"`
	PublicWebURL string `env:"PUBLIC_WEB_URL" envDefault:"http://localhost:3000"`
}

func (a App) IsProd() bool { return a.Env == "production" }

type HTTP struct {
	Host string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" envDefault:"8080"`
}

func (h HTTP) Addr() string { return fmt.Sprintf("%s:%d", h.Host, h.Port) }

type Postgres struct {
	URL string `env:"DATABASE_URL"`
}

type Redis struct {
	URL string `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"`
}

type Auth struct {
	JWTSecret             string        `env:"JWT_SECRET"`
	SessionTTL            time.Duration `env:"SESSION_TTL" envDefault:"720h"`
	APIKeyPrefix          string        `env:"API_KEY_PREFIX" envDefault:"hl_"`
	CalendarEncryptionKey string        `env:"CALENDAR_ENCRYPTION_KEY"`
}

type Email struct {
	ResendAPIKey string `env:"RESEND_API_KEY"`
	From         string `env:"EMAIL_FROM" envDefault:"HaloMail <noreply@halomail.dev>"`
	SMTPHost     string `env:"SMTP_HOST" envDefault:"localhost"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"1025"`
}

// UseResend reports whether a real Resend key is configured; otherwise the
// notification service falls back to local SMTP (Mailpit).
func (e Email) UseResend() bool { return e.ResendAPIKey != "" }

type OAuth struct {
	ClientID     string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
	RedirectURL  string `env:"REDIRECT_URL"`
}

type Microsoft struct {
	ClientID     string `env:"MICROSOFT_CLIENT_ID"`
	ClientSecret string `env:"MICROSOFT_CLIENT_SECRET"`
	TenantID     string `env:"MICROSOFT_TENANT_ID" envDefault:"common"`
	RedirectURL  string `env:"MICROSOFT_REDIRECT_URL"`
}

type OTel struct {
	Enabled     bool    `env:"OTEL_ENABLED" envDefault:"true"`
	Endpoint    string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"http://localhost:4318"`
	ServiceName string  `env:"OTEL_SERVICE_NAME" envDefault:"halomail"`
	SampleRatio float64 `env:"OTEL_TRACES_SAMPLER_ARG" envDefault:"1.0"`
}

type Rate struct {
	PublicRPS   int `env:"RATELIMIT_PUBLIC_RPS" envDefault:"10"`
	PublicBurst int `env:"RATELIMIT_PUBLIC_BURST" envDefault:"20"`
}

// Load reads an optional .env file then parses the environment.
func Load() (*Config, error) {
	_ = godotenv.Load() // best-effort; real env wins
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.Limits.Forms < 0 || cfg.Limits.Meetings < 0 {
		return nil, fmt.Errorf("free limits cannot be negative")
	}
	return cfg, nil
}
