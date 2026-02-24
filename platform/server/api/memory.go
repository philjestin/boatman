package api

import (
	"encoding/json"
	"net/http"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type memoryHandlers struct {
	store storage.MemoryStore
}

func (h *memoryHandlers) listPatterns(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	patterns, err := h.store.ListPatterns(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, patterns)
}

func (h *memoryHandlers) createPattern(w http.ResponseWriter, r *http.Request) {
	var pattern storage.Pattern
	if err := json.NewDecoder(r.Body).Decode(&pattern); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	scope := ScopeFromContext(r.Context())
	if pattern.Scope.OrgID == "" {
		pattern.Scope = scope
	}

	if err := h.store.CreatePattern(r.Context(), &pattern); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, pattern)
}

func (h *memoryHandlers) getPreferences(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	prefs, err := h.store.GetPreferences(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}

func (h *memoryHandlers) setPreferences(w http.ResponseWriter, r *http.Request) {
	var prefs storage.Preferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	scope := ScopeFromContext(r.Context())
	prefs.Scope = scope

	if err := h.store.SetPreferences(r.Context(), &prefs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}
