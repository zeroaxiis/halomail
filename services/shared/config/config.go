
package config

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/shared/usage"
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
	OTPDeliverySecret     string        `env:"OTP_DELIVERY_SECRET"`
	JWTSecret             string        `env:"JWT_SECRET"`
	SessionTTL            time.Duration `env:"SESSION_TTL" envDefault:"720h"`
	APIKeyPrefix          string        `env:"API_KEY_PREFIX" envDefault:"hl_"`
	CalendarEncryptionKey string        `env:"CALENDAR_ENCRYPTION_KEY"`
	RazorpayWebhookSecret string        `env:"RAZORPAY_WEBHOOK_SECRET"`
}

type Email struct {
	ResendAPIKey string `env:"RESEND_API_KEY"`
	From         string `env:"EMAIL_FROM" envDefault:"HaloMail <no-reply@email.zeroaxiis.tech>"`
	SMTPHost     string `env:"SMTP_HOST" envDefault:"localhost"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"1025"`
}

// UseResend reports whether a real Resend key is configured; otherwise the
// notification service falls back to local SMTP (Mailpit).
func (e Email) UseResend() bool {
	return strings.TrimSpace(e.ResendAPIKey) != "" ||
		strings.TrimSpace(os.Getenv("NTS_RESEND_API_KEY")) != "" ||
		strings.TrimSpace(os.Getenv("NOTIFY_RESEND_API_KEY")) != "" ||
		strings.TrimSpace(os.Getenv("RESEND_POOL")) != "" ||
		strings.TrimSpace(os.Getenv("RESEND_KEY_1")) != ""
}

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
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
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

	if strings.TrimSpace(cfg.Auth.CalendarEncryptionKey) == "" {
		h := sha256.Sum256([]byte(cfg.Auth.JWTSecret + ":calendar-encryption-key"))
		cfg.Auth.CalendarEncryptionKey = base64.StdEncoding.EncodeToString(h[:])
	}

	cfg.Email.ResendAPIKey = strings.TrimSpace(strings.Trim(cfg.Email.ResendAPIKey, `"'`))
	if cfg.Email.ResendAPIKey == "" {
		cfg.Email.ResendAPIKey = strings.TrimSpace(strings.Trim(os.Getenv("NTS_RESEND_API_KEY"), `"'`))
	}
	cfg.Email.From = strings.TrimSpace(strings.Trim(cfg.Email.From, `"'`))
	if cfg.Email.From == "" || strings.Contains(cfg.Email.From, "halomail.dev") {
		cfg.Email.From = "HaloMail <no-reply@email.zeroaxiis.tech>"
	}
	if cfg.App.IsProd() {
		if cfg.App.PublicWebURL == "http://localhost:3000" || cfg.App.PublicWebURL == "" {
			cfg.App.PublicWebURL = "https://halomail.zeroaxiis.tech"
		}
		if cfg.App.PublicAPIURL == "http://localhost:8080" || cfg.App.PublicAPIURL == "" {
			cfg.App.PublicAPIURL = "https://halomailapi.zeroaxiis.tech"
		}
	}
	cleanWeb := strings.TrimRight(strings.TrimSpace(cfg.App.PublicWebURL), "/")
	if cleanWeb == "" {
		if cfg.App.IsProd() {
			cleanWeb = "https://halomail.zeroaxiis.tech"
		} else {
			cleanWeb = "http://localhost:3000"
		}
	}
	if strings.TrimSpace(cfg.Google.RedirectURL) == "" || strings.Contains(cfg.Google.RedirectURL, "/v1/oauth/google/callback") || (cfg.App.IsProd() && strings.Contains(cfg.Google.RedirectURL, "localhost")) {
		cfg.Google.RedirectURL = cleanWeb + "/api/calendar/callback"
	}
	return cfg, nil
}
