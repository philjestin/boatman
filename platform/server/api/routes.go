// Package api implements the HTTP API for the Boatman platform.
package api

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/philjestin/boatman-ecosystem/platform/dashboard"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// RegisterRoutes registers all API endpoints on the given mux.
func RegisterRoutes(mux *http.ServeMux, store storage.Store, bus *eventbus.Bus) {
	runs := &runHandlers{store: store.Runs()}
	mem := &memoryHandlers{store: store.Memory()}
	costs := &costHandlers{store: store.Costs()}
	policies := &policyHandlers{store: store.Policies()}
	events := &eventHandlers{store: store.Events(), bus: bus}

	// Wrap all API routes with middleware
	api := http.NewServeMux()

	// Health
	api.HandleFunc("GET /api/v1/health", handleHealth)

	// Runs
	api.HandleFunc("GET /api/v1/runs", runs.list)
	api.HandleFunc("GET /api/v1/runs/{id}", runs.get)
	api.HandleFunc("POST /api/v1/runs", runs.create)

	// Memory
	api.HandleFunc("GET /api/v1/memory/patterns", mem.listPatterns)
	api.HandleFunc("POST /api/v1/memory/patterns", mem.createPattern)
	api.HandleFunc("GET /api/v1/memory/preferences", mem.getPreferences)
	api.HandleFunc("PUT /api/v1/memory/preferences", mem.setPreferences)

	// Costs
	api.HandleFunc("GET /api/v1/costs/summary", costs.summary)
	api.HandleFunc("GET /api/v1/costs/budget", costs.getBudget)
	api.HandleFunc("PUT /api/v1/costs/budget", costs.setBudget)

	// Policies
	api.HandleFunc("GET /api/v1/policies", policies.get)
	api.HandleFunc("PUT /api/v1/policies", policies.set)
	api.HandleFunc("GET /api/v1/policies/effective", policies.getEffective)

	// Events
	api.HandleFunc("GET /api/v1/events", events.query)
	api.HandleFunc("GET /api/v1/events/stream", events.stream)

	// Dashboard (embedded frontend)
	registerDashboard(api)

	// Apply middleware chain
	mux.Handle("/", loggingMiddleware(scopeMiddleware(api)))
}

// registerDashboard serves the embedded web dashboard.
func registerDashboard(mux *http.ServeMux) {
	assets, err := dashboard.Assets()
	if err != nil {
		// Dashboard not available (e.g., dist not built). Serve a fallback.
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("Boatman Platform - Dashboard not built. Run: cd platform/dashboard/frontend && npm run build"))
			} else {
				http.NotFound(w, r)
			}
		})
		return
	}
	fileServer := http.FileServer(http.FS(assets))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// For SPA routing: serve index.html for paths that don't match static files
		if r.URL.Path != "/" {
			_, err := fs.Stat(assets, r.URL.Path[1:])
			if err != nil {
				// Not a static file — serve index.html for client-side routing
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// marshalJSON marshals a value to JSON bytes.
func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
