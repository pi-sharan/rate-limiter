package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort              = "8080"
	defaultBackend               = "memory"
	defaultBucketCapacity        = 5
	defaultRefillRate            = 5
	defaultRefillIntervalSeconds = 60
	defaultRedisAddr             = "127.0.0.1:6379"
	defaultHashReplicas          = 128
	envHTTPPort                  = "HTTP_PORT"
	envBackend                   = "RATE_LIMIT_BACKEND"
	envBucketCapacity            = "RATE_LIMIT_BUCKET_CAPACITY"
	envRefillRate                = "RATE_LIMIT_REFILL_RATE"
	envRefillIntervalSeconds     = "RATE_LIMIT_REFILL_INTERVAL_SECONDS"
	envRedisAddr                 = "REDIS_ADDR"
	envRedisUsername             = "REDIS_USERNAME"
	envRedisPassword             = "REDIS_PASSWORD"
	envRedisDB                   = "REDIS_DB"
	envRedisShards               = "REDIS_SHARDS"
	envRedisHashReplicas         = "REDIS_HASH_REPLICAS"
)

// Load returns a Config populated from environment variables with defaults.
func Load() *Config {
	shards := parseShards()
	if len(shards) == 0 {
		// fall back to a single shard described by the legacy env vars/defaults
		shards = append(shards, RedisShardConfig{
			Addr:     getEnv(envRedisAddr, defaultRedisAddr),
			Username: getEnv(envRedisUsername, ""),
			Password: getEnv(envRedisPassword, ""),
			DB:       getEnvInt(envRedisDB, 0),
		})
	}

	redisCfg := RedisConfig{
		HashReplicas: getEnvInt(envRedisHashReplicas, defaultHashReplicas),
		Shards:       shards,
	}

	return &Config{
		HTTPPort: getEnv(envHTTPPort, defaultHTTPPort),
		RateLimiter: RateLimiterConfig{
			Backend:        getEnv(envBackend, defaultBackend),
			BucketCapacity: getEnvInt(envBucketCapacity, defaultBucketCapacity),
			RefillRate:     getEnvInt(envRefillRate, defaultRefillRate),
			RefillInterval: time.Duration(getEnvInt(envRefillIntervalSeconds, defaultRefillIntervalSeconds)) * time.Second,
		},
		Redis: redisCfg,
	}
}

func parseShards() []RedisShardConfig {
	list := os.Getenv(envRedisShards)
	if list == "" {
		return nil
	}

	username := getEnv(envRedisUsername, "")
	password := getEnv(envRedisPassword, "")
	db := getEnvInt(envRedisDB, 0)

	var shards []RedisShardConfig
	for _, addr := range strings.Split(list, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		shards = append(shards, RedisShardConfig{
			Addr:     addr,
			Username: username,
			Password: password,
			DB:       db,
		})
	}

	return shards
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
