package main

import (
	"log"
	"pz15-vps/internal/config"
	"pz15-vps/internal/server"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting tasks service on port %s", cfg.Port)
	log.Printf("Log level: %s", cfg.LogLevel)

	srv := server.New(cfg)

	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
