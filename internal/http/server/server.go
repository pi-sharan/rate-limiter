package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piyushsharan/rate-limiter/internal/config"
	"github.com/piyushsharan/rate-limiter/internal/http/handlers"
	"github.com/piyushsharan/rate-limiter/internal/http/middleware"
	"github.com/piyushsharan/rate-limiter/internal/limiter"
	redisstore "github.com/piyushsharan/rate-limiter/internal/storage/redis"
	"github.com/piyushsharan/rate-limiter/internal/storage/rediscluster"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
}

// New constructs a server with base middleware and routes.
func New(cfg *config.Config) (*Server, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	limiterSvc, err := buildLimiter(cfg)
	if err != nil {
		return nil, err
	}

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := engine.Group("/")
	protected.Use(middleware.RateLimiter(limiterSvc))
	protected.GET("/resource", handlers.Resource)

	return &Server{
		engine: engine,
		cfg:    cfg,
	}, nil
}

func (s *Server) Run() error {
	return s.engine.Run(fmt.Sprintf(":%s", s.cfg.HTTPPort))
}

func buildLimiter(cfg *config.Config) (limiter.Limiter, error) {
	base := limiter.TokenBucketConfig{
		BucketCapacity: cfg.RateLimiter.BucketCapacity,
		RefillRate:     cfg.RateLimiter.RefillRate,
		RefillInterval: cfg.RateLimiter.RefillInterval,
	}

	switch cfg.RateLimiter.Backend {
	case "redis":
		return buildRedisLimiter(cfg, base)
	default:
		return limiter.NewInMemory(base), nil
	}
}

func buildRedisLimiter(cfg *config.Config, base limiter.TokenBucketConfig) (limiter.Limiter, error) {
	if len(cfg.Redis.Shards) == 0 {
		return nil, fmt.Errorf("redis backend selected but no shards configured")
	}

	nodes := make([]rediscluster.Node, 0, len(cfg.Redis.Shards))
	for i, shard := range cfg.Redis.Shards {
		client, err := redisstore.New(redisstore.Config{
			Addr:     shard.Addr,
			Username: shard.Username,
			Password: shard.Password,
			DB:       shard.DB,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create redis client for %d (%s): %w", i, shard.Addr, err)
		}
		nodes = append(nodes, rediscluster.Node{
			ID:     fmt.Sprintf("redis-%d-%s", i, shard.Addr),
			Client: client.Raw(),
		})
	}

	ring, err := rediscluster.NewConsistentHash(nodes, cfg.Redis.HashReplicas)
	if err != nil {
		return nil, err
	}

	return limiter.NewRedisWithPicker(ring, base)
}
