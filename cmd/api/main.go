package main

import (
	"log"

	"github.com/piyushsharan/rate-limiter/internal/config"
	"github.com/piyushsharan/rate-limiter/internal/http/server"
)

func main() {
	cfg := config.Load()
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
