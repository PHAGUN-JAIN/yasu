package handler

import (
	"encoding/json"
	"net/http"
	"urlshort/internal/service"
	"urlshort/internal/utils/serde"
)

type YasuHandler struct {
	Service *service.YasuService
}

// Get returns the shortened URL for a given original URL
func (h *YasuHandler) Get(w http.ResponseWriter, r *http.Request) {
	msg := h.Service.Get()
	w.Write([]byte(msg))
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
	err = serializer.YasuDeserializer(req)
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
