package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/piyushsharan/rate-limiter/internal/config"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
}

// New constructs a server with base middleware and routes.
func New(cfg *config.Config) *Server {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// Placeholder route
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{
		engine: engine,
		cfg:    cfg,
	}
}

func (s *Server) Run() error {
	return s.engine.Run(fmt.Sprintf(":%s", s.cfg.HTTPPort))
}
