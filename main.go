package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Room struct {
	Name    string
	Clients map[*websocket.Conn]bool
}

var rooms = make(map[string]Room)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
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
	http.HandleFunc("/ws", wsHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("WebSocket server started on http://localhost:%s/ws\n", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"Hello": "World"})
	})
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
