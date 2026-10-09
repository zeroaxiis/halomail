package backup

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Enabled       bool          `env:"R2_BACKUP_ENABLED" envDefault:"true"`
	Endpoint      string        `env:"R2_ENDPOINT"`
	AccessKey     string        `env:"R2_ACCESS_KEY_ID"`
	SecretKey     string        `env:"R2_SECRET_ACCESS_KEY"`
	Bucket        string        `env:"R2_BUCKET_NAME"`
	Region        string        `env:"R2_REGION" envDefault:"auto"`
	Prefix        string        `env:"R2_BACKUP_PREFIX" envDefault:"halomail/production"`
	DatabaseURL   string        `env:"BACKUP_DATABASE_URL"`
	Interval      time.Duration `env:"R2_BACKUP_INTERVAL" envDefault:"1h"`
	Timeout       time.Duration `env:"R2_BACKUP_TIMEOUT" envDefault:"15m"`
	MaxAge        time.Duration `env:"R2_BACKUP_MAX_AGE" envDefault:"2h"`
	DumpBinary    string        `env:"PG_DUMP_PATH" envDefault:"pg_dump"`
	RestoreBinary string        `env:"PG_RESTORE_PATH" envDefault:"pg_restore"`
}

// Load uses the already-loaded server environment. Credentials never enter logs.
func Load(databaseURL string) (Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return c, fmt.Errorf("invalid R2 backup configuration")
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = databaseURL
	}
	return c, c.Validate()
}

func (c Config) Active() bool {
	return c.Enabled && (c.Endpoint != "" || c.AccessKey != "" || c.SecretKey != "" || c.Bucket != "")
}

func (c Config) Validate() error {
	if !c.Active() {
		return nil
	}
	if c.Endpoint == "" || c.AccessKey == "" || c.SecretKey == "" || c.Bucket == "" {
		return fmt.Errorf("R2 backup requires R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY and R2_BUCKET_NAME")
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("R2_ENDPOINT must be an HTTPS S3 endpoint without credentials, bucket path or query parameters")
	}
	if c.Prefix == "" || strings.HasPrefix(c.Prefix, "/") || path.Clean(c.Prefix) != c.Prefix || strings.Contains(c.Prefix, "\\") || c.Prefix == "." || c.Prefix == ".." || strings.HasPrefix(c.Prefix, "../") {
		return fmt.Errorf("R2_BACKUP_PREFIX must be a nonempty object prefix without traversal")
	}
	if c.Interval < time.Minute || c.Timeout <= 0 || c.MaxAge <= c.Interval+c.Timeout {
		return fmt.Errorf("R2_BACKUP_INTERVAL must be at least 1m; timeout must be positive; max age must exceed interval plus timeout")
	}
	return nil
}
