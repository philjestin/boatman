package api

import (
	"encoding/json"
	"net/http"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type policyHandlers struct {
	store storage.PolicyStore
}

func (h *policyHandlers) get(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	policy, err := h.store.Get(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if policy == nil {
		writeJSON(w, http.StatusOK, map[string]any{"policy": nil})
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

func (h *policyHandlers) set(w http.ResponseWriter, r *http.Request) {
	var policy storage.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	scope := ScopeFromContext(r.Context())
	policy.Scope = scope

	if err := h.store.Set(r.Context(), &policy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

func (h *policyHandlers) getEffective(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	policy, err := h.store.GetEffectivePolicy(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if policy == nil {
		writeJSON(w, http.StatusOK, map[string]any{"policy": nil})
		return
	}

	writeJSON(w, http.StatusOK, policy)
}
