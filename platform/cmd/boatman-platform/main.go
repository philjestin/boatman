// Command boatman-platform runs the Boatman organizational platform server.
//
// Usage:
//
//	boatman-platform --data-dir ~/.boatman/platform --port 8080
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/server"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	dataDir := flag.String("data-dir", defaultDataDir(), "Data directory for SQLite and NATS")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	// Initialize storage
	dbPath := filepath.Join(*dataDir, "platform.db")
	store, err := sqlite.New(sqlite.Options{Path: dbPath})
	if err != nil {
		log.Fatalf("create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Initialize event bus
	bus, err := eventbus.New(
		eventbus.WithPort(-1), // embedded, pick random port
		eventbus.WithDataDir(filepath.Join(*dataDir, "nats")),
		eventbus.WithEventStore(store.Events()),
	)
	if err != nil {
		log.Fatalf("create event bus: %v", err)
	}
	defer bus.Close()

	// Start server
	srv := server.New(server.Config{
		Port:  *port,
		Store: store,
		Bus:   bus,
	})

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10_000_000_000) // 10 seconds
		defer cancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	fmt.Printf("Boatman Platform starting on port %d (data: %s)\n", *port, *dataDir)
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".boatman", "platform")
}
