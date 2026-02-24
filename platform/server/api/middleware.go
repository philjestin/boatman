package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type contextKey string

const scopeKey contextKey = "scope"

// ScopeFromContext retrieves the Scope from the request context.
func ScopeFromContext(ctx context.Context) storage.Scope {
	if s, ok := ctx.Value(scopeKey).(storage.Scope); ok {
		return s
	}
	return storage.Scope{}
}

// scopeMiddleware extracts org/team headers into context.
func scopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := storage.Scope{
			OrgID:  r.Header.Get("X-Boatman-Org"),
			TeamID: r.Header.Get("X-Boatman-Team"),
			RepoID: r.Header.Get("X-Boatman-Repo"),
		}
		ctx := context.WithValue(r.Context(), scopeKey, scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware logs incoming requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
