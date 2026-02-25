package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Room struct {
	Name    string
	Clients map[*websocket.Conn]bool
}

var rooms = make(map[string]Room)

func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		return []string{"*"}
	}
	return strings.Split(origins, ",")
}

func isOriginAllowed(origin string) bool {
	allowed := getAllowedOrigins()
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(getAllowedOrigins()) == 1 && getAllowedOrigins()[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return isOriginAllowed(r.Header.Get("Origin"))
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id query parameter", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	go handleConnection(conn, id)
}

func closeConnectionsInRoom(id string) {
	room := rooms[id]
	for client := range room.Clients {
		client.Close()
	}
	delete(rooms, id)
	fmt.Printf("Current rooms after removal: %v\n", rooms)
}

func handleConnection(conn *websocket.Conn, id string) {
	defer closeConnectionsInRoom(id)

	currRoom, alreadyExists := rooms[id]

	if !alreadyExists {
		fmt.Println("Opening new room")
		newRoom := Room{
			Name:    fmt.Sprintf("room%d", len(rooms)+1),
			Clients: make(map[*websocket.Conn]bool),
		}
		rooms[id] = newRoom
		currRoom = newRoom
	}

	if len(currRoom.Clients) >= 2 {
		fmt.Printf("Room %s is full, closing connection\n", currRoom.Name)
		conn.WriteMessage(websocket.TextMessage, []byte("Room is full"))
		conn.Close()
		return
	}

	currRoom.Clients[conn] = true

	fmt.Printf("Current rooms: %v\n", rooms)

	fmt.Println("Waiting for message...")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received: %s\n", message)

		for client := range currRoom.Clients {
			if client != conn {
				fmt.Printf("Sending message to client %p\n", client.RemoteAddr())
				if err = client.WriteMessage(websocket.TextMessage, message); err != nil {
					fmt.Printf("Error sending message to client %p: %v\n", client.RemoteAddr(), err)
					break
				}
			}
		}
	}
}

func main() {
	fmt.Println("Running...")
	http.HandleFunc("/ws", corsMiddleware(wsHandler))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Allowed origins: %v\n", getAllowedOrigins())
	fmt.Printf("WebSocket server started on http://localhost:%s/ws\n", port)
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"Hello": "World"})
	}))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
