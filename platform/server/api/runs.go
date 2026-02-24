package api

import (
	"encoding/json"
	"net/http"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type runHandlers struct {
	store storage.RunStore
}

func (h *runHandlers) list(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	filter := storage.RunFilter{}
	if scope.OrgID != "" || scope.TeamID != "" || scope.RepoID != "" {
		filter.Scope = &scope
	}

	// Parse query params
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = storage.RunStatus(status)
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filter.UserID = userID
	}

	runs, err := h.store.List(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, runs)
}

func (h *runHandlers) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing run id", http.StatusBadRequest)
		return
	}

	run, err := h.store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (h *runHandlers) create(w http.ResponseWriter, r *http.Request) {
	var run storage.Run
	if err := json.NewDecoder(r.Body).Decode(&run); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Apply scope from headers if not set in body
	scope := ScopeFromContext(r.Context())
	if run.Scope.OrgID == "" {
		run.Scope.OrgID = scope.OrgID
	}
	if run.Scope.TeamID == "" {
		run.Scope.TeamID = scope.TeamID
	}

	if err := h.store.Create(r.Context(), &run); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, run)
}
