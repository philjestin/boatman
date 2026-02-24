// Package server provides the HTTP server for the Boatman platform.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/server/api"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Server is the Boatman platform HTTP server.
type Server struct {
	httpServer *http.Server
	store      storage.Store
	bus        *eventbus.Bus
	addr       string
}

// Config configures the server.
type Config struct {
	Port    int
	Store   storage.Store
	Bus     *eventbus.Bus
}

// New creates a new Server.
func New(cfg Config) *Server {
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, cfg.Store, cfg.Bus)

	addr := fmt.Sprintf(":%d", cfg.Port)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 0, // SSE needs no write timeout
			IdleTimeout:  120 * time.Second,
		},
		store: cfg.Store,
		bus:   cfg.Bus,
		addr:  addr,
	}
}

// Start begins listening. It blocks until the server is stopped.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.addr = ln.Addr().String()
	log.Printf("platform server listening on %s", s.addr)
	return s.httpServer.Serve(ln)
}

// Addr returns the address the server is listening on.
func (s *Server) Addr() string {
	return s.addr
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
