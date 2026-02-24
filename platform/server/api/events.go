package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type eventHandlers struct {
	store storage.EventStore
	bus   *eventbus.Bus
}

func (h *eventHandlers) query(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	filter := storage.EventFilter{}
	if scope.OrgID != "" || scope.TeamID != "" {
		filter.Scope = &scope
	}

	if runID := r.URL.Query().Get("run_id"); runID != "" {
		filter.RunID = runID
	}
	if types := r.URL.Query().Get("types"); types != "" {
		filter.Types = strings.Split(types, ",")
	}
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.Since = t
		}
	}

	events, err := h.store.Query(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, events)
}

// stream handles SSE connections for real-time event streaming.
func (h *eventHandlers) stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	scope := ScopeFromContext(r.Context())

	// Build NATS subject from scope
	subject := eventbus.AllEventsSubject
	if scope.OrgID != "" && scope.TeamID != "" {
		subject = eventbus.TeamWildcard(scope.OrgID, scope.TeamID)
	} else if scope.OrgID != "" {
		subject = eventbus.OrgWildcard(scope.OrgID)
	}

	// Subscribe to events
	ch, cancel, err := h.bus.Subscribe(r.Context(), subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cancel()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := marshalJSON(event)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
