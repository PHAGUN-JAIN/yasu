package http

import (
	"context"
	"log"
	"net/http"
	"time"
	"urlshort/cmd/urlshort/initialize"
	"urlshort/internal/http/handler"
	"urlshort/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func RunYasuServer(port string) error {

	mainRouter := mux.NewRouter()
	mainRouter.Use(RequestIDMiddleware)
	mainRouter.Use(LoggingMiddleware)
	apiGroup := mainRouter.PathPrefix("/api/v1").Subrouter()

	RegisterHealthRoute(apiGroup)
	RegisterYasuRoutes(apiGroup)

	srv := &http.Server{
		Handler: mainRouter,
		Addr:    ":" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(srv.ListenAndServe())

	return nil
}

func RegisterYasuRoutes(router *mux.Router) {
	// Initialize the YasuHandler with the YasuService, postgress client , buffer client, and any other dependencies it may require
	pgconn := initialize.GetPostgressClient()
	yasuHandler := &handler.YasuHandler{Service: &service.YasuService{
		PgConn: pgconn,
	}}

	router.HandleFunc("/{code}", yasuHandler.Get).Methods("GET")
	router.HandleFunc("/shorten", yasuHandler.Create).Methods("POST")
	router.HandleFunc("/{code}", yasuHandler.Delete).Methods("DELETE")
}

func RegisterHealthRoute(router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	},
	).Methods("GET")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		reqID, _ := r.Context().Value("request_id").(string)

		log.Printf("[%s] Started %s %s",
			reqID, r.Method, r.URL.Path,
		)

		next.ServeHTTP(w, r)

		log.Printf("[%s] Completed in %v",
			reqID, time.Since(start),
		)
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		// attach to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
