package handler

import (
	"net/http"
)

func YasuHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintln(w, "Hello from YasuHandler 🚀")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Hello from YasuHandler"}`))
}
