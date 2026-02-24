package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type costHandlers struct {
	store storage.CostStore
}

func (h *costHandlers) summary(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	group := storage.TimeGroup(r.URL.Query().Get("group"))
	if group == "" {
		group = storage.TimeGroupDay
	}

	since := time.Now().AddDate(0, -1, 0) // default: last month
	until := time.Now()

	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}
	if u := r.URL.Query().Get("until"); u != "" {
		if t, err := time.Parse(time.RFC3339, u); err == nil {
			until = t
		}
	}

	summaries, err := h.store.GetUsageSummary(r.Context(), scope, group, since, until)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, summaries)
}

func (h *costHandlers) getBudget(w http.ResponseWriter, r *http.Request) {
	scope := ScopeFromContext(r.Context())

	budget, err := h.store.GetBudget(r.Context(), scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if budget == nil {
		writeJSON(w, http.StatusOK, map[string]any{"budget": nil})
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

func (h *costHandlers) setBudget(w http.ResponseWriter, r *http.Request) {
	var budget storage.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	scope := ScopeFromContext(r.Context())
	budget.Scope = scope

	if err := h.store.SetBudget(r.Context(), &budget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, budget)
}
