package handler

import (
	"encoding/json"
	"net/http"
	"urlshort/internal/service"
)

type YasuHandler struct {
	Service *service.YasuService
}

// GetYasu returns the shortened URL for a given original URL
func (h *YasuHandler) GetYasu(w http.ResponseWriter, r *http.Request) {
	msg := h.Service.GetMessage()
	w.Write([]byte(msg))
}

// CreateYasu creates a new shortened URL for a given original URL
func (h *YasuHandler) CreateYasu(w http.ResponseWriter, r *http.Request) {
	// validation, auth checks, etc. can be done here before calling the service layer
	var req service.ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req) // TODO: handle error and validate req
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.Service.Shorten(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}

func (h *YasuHandler) DeleteYasu(w http.ResponseWriter, r *http.Request) {
	// validation, auth checks, etc. can be done here before calling the service layer
	result := h.Service.DeleteYasu()
	w.Write([]byte(result))
}
