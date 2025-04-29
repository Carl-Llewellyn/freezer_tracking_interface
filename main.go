package main

import (
	"log"
	"net/http"
)

func main() {
	// Allow CORS for all origins and methods (for quick prototyping)
	handler := http.FileServer(http.Dir("static"))
	corsHandler := func(w http.ResponseWriter, r *http.Request) {
		// Allow everything for prototype purposes
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		handler.ServeHTTP(w, r)
	}

	http.HandleFunc("/", corsHandler)

	log.Println("Serving static/ on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
