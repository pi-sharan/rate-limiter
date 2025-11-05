package main

import (
	"log"

	"github.com/piyushsharan/rate-limiter/internal/config"
	"github.com/piyushsharan/rate-limiter/internal/http/server"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
