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
		client, err := redisstore.New(redisstore.Config{
			Addr:     cfg.Redis.Addr,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err != nil {
			return nil, err
		}

		return limiter.NewRedis(client.Raw(), base)
	default:
		return limiter.NewInMemory(base), nil
	}
}
