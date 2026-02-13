package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr              string
	DBURL                 string
	RedisURL              string
	NATSURL               string
	JWTIssuer             string
	JWTAudience           string
	JWTKeyCurrent         string
	JWTKeyPrevious        string
	FieldKeyCurrent       string
	FieldKeyPrevious      string
	SearchKey             string
	CorsOrigins           []string
	RateLimitAPIPerMin    int
	RateLimitLoginPerMin  int
	BodyMaxBytes          int64
	AnalyticsServiceToken string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		DBURL:                 fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", getEnv("APP_DB_USER", "erp_app"), getEnv("APP_DB_PASSWORD", "change-me-app"), getEnv("POSTGRES_HOST", "localhost"), getEnv("POSTGRES_PORT", "5432"), getEnv("POSTGRES_DB", "erpwms")),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		NATSURL:               getEnv("NATS_URL", "nats://localhost:4222"),
		JWTIssuer:             getEnv("JWT_ISSUER", "erpwms"),
		JWTAudience:           getEnv("JWT_AUDIENCE", "erpwms-users"),
		JWTKeyCurrent:         os.Getenv("JWT_SIGNING_KEY_CURRENT"),
		JWTKeyPrevious:        os.Getenv("JWT_SIGNING_KEY_PREVIOUS"),
		FieldKeyCurrent:       os.Getenv("FIELD_ENC_MASTER_KEY_CURRENT"),
		FieldKeyPrevious:      os.Getenv("FIELD_ENC_MASTER_KEY_PREVIOUS"),
		SearchKey:             os.Getenv("SEARCH_KEY"),
		CorsOrigins:           strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:8080"), ","),
		RateLimitAPIPerMin:    getEnvInt("RATE_LIMIT_API_PER_MIN", 120),
		RateLimitLoginPerMin:  getEnvInt("RATE_LIMIT_LOGIN_PER_MIN", 10),
		BodyMaxBytes:          int64(getEnvInt("BODY_MAX_BYTES", 1048576)),
		AnalyticsServiceToken: getEnv("ANALYTICS_SERVICE_TOKEN", "replace-token"),
	}
	if cfg.JWTKeyCurrent == "" || cfg.FieldKeyCurrent == "" || cfg.SearchKey == "" {
		return cfg, fmt.Errorf("missing required security keys")
	}
	return cfg, nil
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getEnvInt(k string, d int) int {
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
