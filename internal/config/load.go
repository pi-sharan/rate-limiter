package config

import "os"

const defaultHTTPPort = "8080"

// Load returns a Config populated from environment variables with defaults.
func Load() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", defaultHTTPPort),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
