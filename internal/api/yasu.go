package http

import (
	"context"
	"log"
	"net/http"
	"time"
	"urlshort/cmd/initialize"
	"urlshort/internal/api/handler"
	"urlshort/internal/service/outbox"
	urlshortsvc "urlshort/internal/service/urlshort"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RunYasuServer(port string) error {
	clients := initialize.GetClients()

	worker := outbox.NewWorker(clients.PgClient)
	drainer := outbox.NewDrainer(clients.PgClient, clients.ChClient)

	mainRouter := mux.NewRouter()
	mainRouter.Use(RequestIDMiddleware)
	mainRouter.Use(LoggingMiddleware)
	apiGroup := mainRouter.PathPrefix("/api/v1").Subrouter()

	RegisterHealthRoute(apiGroup)
	RegisterYasuRoutes(apiGroup, clients, worker)
	RegisterAnalyticsRoutes(apiGroup, clients)

	// start both as goroutines
	go worker.Start(context.Background())
	go drainer.Start(context.Background())

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

func RegisterHealthRoute(router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	},
	).Methods("GET")
}

func RegisterYasuRoutes(router *mux.Router, clients *initialize.Clients, worker *outbox.Worker) {
	// Initialize the YasuHandler with the YasuService, postgress client , buffer client, and any other dependencies it may require
	yasuHandler := &handler.YasuHandler{Service: urlshortsvc.NewYasuService(clients.PgClient, clients.Cache, worker)}

	router.HandleFunc("/{code}", yasuHandler.Get).Methods("GET")
	router.HandleFunc("/shorten", yasuHandler.Create).Methods("POST")
	router.HandleFunc("/{code}", yasuHandler.Delete).Methods("DELETE")
}

func RegisterAnalyticsRoutes(router *mux.Router, clients *initialize.Clients) {
	analyticsHandler := &handler.AnalyticsHandler{Service: urlshortsvc.NewAnalyticsService(clients.ChClient)}

	router.HandleFunc("/analytics/{code}", analyticsHandler.Get).Methods("GET")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		reqID, _ := r.Context().Value(RequestIDKey).(string)

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
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
