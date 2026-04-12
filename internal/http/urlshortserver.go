package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func RegisterHealthRoute(router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}

func RunServer(port string) error {

	mainRouter := mux.NewRouter()
	apiGroup := mainRouter.PathPrefix("/api/v1").Subrouter()

	RegisterHealthRoute(apiGroup)

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
