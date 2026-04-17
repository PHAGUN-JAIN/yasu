package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"urlshort/internal/service"
	"urlshort/internal/utils/serde"

	"github.com/gorilla/mux"
)

type YasuHandler struct {
	Service *service.YasuService
}

// Get returns the shortened URL for a given original URL
func (h *YasuHandler) Get(w http.ResponseWriter, r *http.Request) {

	serializer := &serde.YasuSerializer{}
	err := serializer.ResolveDeserializer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	vars := mux.Vars(r)
	code := vars["code"]
	log.Printf("extracted code: '%s'", code)

	originalURL, err := h.Service.Get(r.Context(), code)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "URL not found or expired", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error retrieving original URL for code %s: %v", code, err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	// 302 — not 301, see design decisions in README
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Create creates a new shortened URL for a given original URL
//
//	curl -X POST http://localhost:8000/api/v1/shorten \
//	  -H "Content-Type: application/json" \
//	  -d '{"url": "https://google.com"}' \
//	  -v
func (h *YasuHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req) // TODO: handle error and validate req
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	serializer := &serde.YasuSerializer{}
	err = serializer.ShortenDeserializer(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, err := h.Service.Shorten(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"short_url": "http://localhost:8000/" + code,
		"code":      code,
	})
}

func (h *YasuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// validation, auth checks, etc. can be done here before calling the service layer
	result := h.Service.Delete()
	w.Write([]byte(result))
}
