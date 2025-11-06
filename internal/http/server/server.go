package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piyushsharan/rate-limiter/internal/config"
	"github.com/piyushsharan/rate-limiter/internal/http/handlers"
	"github.com/piyushsharan/rate-limiter/internal/http/middleware"
	"github.com/piyushsharan/rate-limiter/internal/limiter"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
}

// New constructs a server with base middleware and routes.
func New(cfg *config.Config) *Server {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	limiterCfg := limiter.TokenBucketConfig{
		BucketCapacity: cfg.RateLimiter.BucketCapacity,
		RefillRate:     cfg.RateLimiter.RefillRate,
		RefillInterval: cfg.RateLimiter.RefillInterval,
	}
	limiterSvc := limiter.NewInMemory(limiterCfg)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := engine.Group("/")
	protected.Use(middleware.RateLimiter(limiterSvc))
	protected.GET("/resource", handlers.Resource)

	return &Server{
		engine: engine,
		cfg:    cfg,
	}
}

func (s *Server) Run() error {
	return s.engine.Run(fmt.Sprintf(":%s", s.cfg.HTTPPort))
}
