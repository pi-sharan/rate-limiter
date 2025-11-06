package config

import "time"

type Config struct {
	HTTPPort    string
	RateLimiter RateLimiterConfig
}

type RateLimiterConfig struct {
	BucketCapacity int
	RefillRate     int
	RefillInterval time.Duration
}
