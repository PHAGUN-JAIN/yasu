package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	ob "urlshort/internal/service/outbox"
	urlshortsvc "urlshort/internal/service/urlshort"
	"urlshort/internal/utils/serde"

	"github.com/gorilla/mux"
)

type YasuHandler struct {
	Service *urlshortsvc.YasuService
}

// Get returns the shortened URL for a given original URL
func (h *YasuHandler) Get(w http.ResponseWriter, r *http.Request) {
	//  1. extract code from URL ✅
	//  2. check Redis cache first ✅
	//  3. if cache miss → query Postgres ✅
	//  4. if not found → 404    ✅
	//  5. if found but expired → 410 Gone ✅
	//  6. redirect 302 to original URL ✅

	serializer := &serde.YasuSerializer{}
	err := serializer.ResolveDeserializer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	vars := mux.Vars(r)
	code := vars["code"]
	log.Printf("extracted code: '%s'", code)

	originalURL, err := h.Service.Get(r.Context(), code, r.Header, ob.SourceREST)
	if errors.Is(err, urlshortsvc.ErrNotFound) {
		http.Error(w, urlshortsvc.ErrNotFound.Error(), http.StatusNotFound)
		return
	}
	if errors.Is(err, urlshortsvc.ErrExpired) {
		http.Error(w, urlshortsvc.ErrExpired.Error(), http.StatusGone)
		return
	}
	if err != nil {
		http.Error(w, urlshortsvc.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	// 302 — not 301, see design decisions in README
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Create creates a new shortened URL for a given original URL
func (h *YasuHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req urlshortsvc.ShortenRequest
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
		http.Error(w, urlshortsvc.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	// return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"short_url": "http://localhost:8000/" + code,
		"code":      code,
	}); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *YasuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// validation, auth checks, etc. can be done here before calling the service layer
	result := h.Service.Delete()
	if _, err := w.Write([]byte(result)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
