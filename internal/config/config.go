package config

import "time"

type Config struct {
	HTTPPort    string
	RateLimiter RateLimiterConfig
	Redis       RedisConfig
}

type RateLimiterConfig struct {
	Backend        string
	BucketCapacity int
	RefillRate     int
	RefillInterval time.Duration
}

type RedisConfig struct {
	HashReplicas int
	Shards       []RedisShardConfig
}

type RedisShardConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}
