package main

import (
	"log"
	"net/http"
)

func main() {
	config := DefaultConfig()
	hub := NewHub(config)
	go hub.Run()

	log.Printf("Starting server on port %s", config.Port)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	log.Fatal(http.ListenAndServe(config.Port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Server is running"	}`))
}
