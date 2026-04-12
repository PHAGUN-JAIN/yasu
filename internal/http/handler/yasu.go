package handler

import (
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
	result := h.Service.Shorten(r)
	w.Write([]byte(result))
}

func (h *YasuHandler) DeleteYasu(w http.ResponseWriter, r *http.Request) {
	// validation, auth checks, etc. can be done here before calling the service layer
	result := h.Service.DeleteYasu()
	w.Write([]byte(result))
}
