package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/create-room", corsMiddleware(createRoomHandler))
	mux.HandleFunc("/ws", corsMiddleware(wsHandler))
	mux.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"Hello": "World"})
	}))
	return mux
}

func runServer() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Allowed origins: %v\n", getAllowedOrigins())
	fmt.Printf("WebSocket server started on http://localhost:%s/ws\n", port)
	return http.ListenAndServe(":"+port, newHandler())
}
