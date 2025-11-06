package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPPort              = "8080"
	defaultBucketCapacity        = 5
	defaultRefillRate            = 5
	defaultRefillIntervalSeconds = 60
	envBucketCapacity            = "RATE_LIMIT_BUCKET_CAPACITY"
	envRefillRate                = "RATE_LIMIT_REFILL_RATE"
	envRefillIntervalSeconds     = "RATE_LIMIT_REFILL_INTERVAL_SECONDS"
	envHTTPPort                  = "HTTP_PORT"
)

// Load returns a Config populated from environment variables with defaults.
func Load() *Config {
	return &Config{
		HTTPPort: getEnv(envHTTPPort, defaultHTTPPort),
		RateLimiter: RateLimiterConfig{
			BucketCapacity: getEnvInt(envBucketCapacity, defaultBucketCapacity),
			RefillRate:     getEnvInt(envRefillRate, defaultRefillRate),
			RefillInterval: time.Duration(getEnvInt(envRefillIntervalSeconds, defaultRefillIntervalSeconds)) * time.Second,
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
