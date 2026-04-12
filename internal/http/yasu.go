package http

import (
	"log"
	"net/http"
	"time"
	"urlshort/internal/http/handler"
	"urlshort/internal/service"

	"github.com/gorilla/mux"
)

func RunYasuServer(port string) error {

	mainRouter := mux.NewRouter()
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
	yasuHandler := &handler.YasuHandler{Service: &service.YasuService{}} // initialize the service

	router.HandleFunc("/yasu/{id}", yasuHandler.GetYasu).Methods("GET")
	router.HandleFunc("/yasu/shorten", yasuHandler.CreateYasu).Methods("POST")
	router.HandleFunc("/yasu/{id}", yasuHandler.DeleteYasu).Methods("DELETE")
}

func RegisterHealthRoute(router *mux.Router) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	},
	).Methods("GET")
}
