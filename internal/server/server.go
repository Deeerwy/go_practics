package server

import (
	"fmt"
	"log"
	"net/http"
	"pz15-vps/internal/config"
	"pz15-vps/internal/handler"
)

type Server struct {
	cfg *config.Config
	mux *http.ServeMux
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", handler.Health)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	log.Printf("HTTP server listening on %s", addr)

	return http.ListenAndServe(addr, s.mux)
}
