package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                  string
	HTTPAddr             string
	DBURL                string
	RedisAddr            string
	NATSURL              string
	JWTIssuer            string
	JWTAudience          string
	JWTCurrent           string
	JWTPrevious          string
	SearchPepper         string
	AuditPepper          string
	FieldEncCurrentB64   string
	FieldEncPreviousB64  string
	FieldEncCurrentKeyID string
	FieldEncPrevKeyID    string
	CorsOrigins          []string
	CookieSecure         bool
	RateLimitLogin       int
	RateLimitAPI         int
	AutotestEnabled      bool
	AutotestToken        string
}

func Load() (Config, error) {
	env := get("ENV", "dev")
	cookieSecure := getBool("COOKIE_SECURE", env == "prod")
	cors := strings.Split(get("CORS_ALLOWED_ORIGINS", "http://localhost:8080"), ",")
	cfg := Config{
		Env:                  env,
		HTTPAddr:             get("HTTP_ADDR", ":8080"),
		DBURL:                get("DB_URL", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", get("APP_DB_USER", "erp_app"), get("APP_DB_PASSWORD", "change-me-app"), get("POSTGRES_HOST", "localhost"), get("POSTGRES_PORT", "5432"), get("POSTGRES_DB", "erpwms"))),
		RedisAddr:            get("REDIS_ADDR", "redis:6379"),
		NATSURL:              get("NATS_URL", "nats://nats:4222"),
		JWTIssuer:            get("JWT_ISSUER", "erpwms"),
		JWTAudience:          get("JWT_AUDIENCE", "erpwms-users"),
		JWTCurrent:           os.Getenv("JWT_SIGNING_KEY_CURRENT"),
		JWTPrevious:          os.Getenv("JWT_SIGNING_KEY_PREVIOUS"),
		SearchPepper:         os.Getenv("SEARCH_PEPPER"),
		AuditPepper:          os.Getenv("AUDIT_PEPPER"),
		FieldEncCurrentB64:   os.Getenv("FIELD_ENC_MASTER_KEY_CURRENT"),
		FieldEncPreviousB64:  os.Getenv("FIELD_ENC_MASTER_KEY_PREVIOUS"),
		FieldEncCurrentKeyID: get("FIELD_ENC_KEY_ID_CURRENT", "v1"),
		FieldEncPrevKeyID:    get("FIELD_ENC_KEY_ID_PREVIOUS", "v0"),
		CorsOrigins:          cors,
		CookieSecure:         cookieSecure,
		RateLimitLogin:       getInt("RATE_LIMIT_LOGIN_PER_MIN", 10),
		RateLimitAPI:         getInt("RATE_LIMIT_API_PER_MIN", 120),
		AutotestEnabled:      getBool("AUTOTEST_ENABLED", false),
		AutotestToken:        os.Getenv("AUTOTEST_TOKEN"),
	}
	if cfg.Env == "prod" {
		if !cfg.CookieSecure {
			return cfg, fmt.Errorf("COOKIE_SECURE must be true in prod")
		}
		for _, o := range cfg.CorsOrigins {
			if strings.TrimSpace(o) == "*" {
				return cfg, fmt.Errorf("wildcard cors forbidden in prod")
			}
		}
		if cfg.JWTCurrent == "" || cfg.SearchPepper == "" || cfg.AuditPepper == "" || cfg.FieldEncCurrentB64 == "" {
			return cfg, fmt.Errorf("missing required security keys")
		}
	}
	return cfg, nil
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}
func getBool(k string, d bool) bool {
	v := strings.ToLower(os.Getenv(k))
	if v == "true" || v == "1" {
		return true
	}
	if v == "false" || v == "0" {
		return false
	}
	return d
}
