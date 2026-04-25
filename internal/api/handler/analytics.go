package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	urlshortsvc "urlshort/internal/service/urlshort"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	Service *urlshortsvc.AnalyticsService
}

// Get returns analytics data for a given shortened URL code
func (h *AnalyticsHandler) Get(w http.ResponseWriter, r *http.Request) {
	// 1. extract code from URL
	// 2. query ClickHouse for analytics data
	// 3. return analytics data as JSON

	vars := mux.Vars(r)
	code := vars["code"]
	log.Printf("extracted code for analytics: '%s'", code)

	analyticsData, err := h.Service.Get(r.Context(), code)
	if errors.Is(err, urlshortsvc.ErrNotFound) {
		http.Error(w, urlshortsvc.ErrNotFound.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, urlshortsvc.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analyticsData)
}
