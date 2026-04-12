package http

import (
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
)

func RunServer(port string) error{

	r := mux.NewRouter()

	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		http.ResponseWriter.Write(w, []byte("Hello, World!"))
	},
	).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    ":" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(srv.ListenAndServe())

	return nil
}
