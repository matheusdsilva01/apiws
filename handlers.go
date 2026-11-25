package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return isOriginAllowed(r.Header.Get("Origin"))
	},
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		http.Error(w, "Could not create room", http.StatusInternalServerError)
		return
	}
	id := hex.EncodeToString(idBytes[:])

	rooms[id] = Room{
		Name:      fmt.Sprintf("room%d", len(rooms)+1),
		Clients:   make(map[*websocket.Conn]bool),
		ExpiresAt: time.Now().Add(roomTTL),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"id": id}); err != nil {
		fmt.Println("Error writing create-room response:", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id query parameter", http.StatusBadRequest)
		return
	}

	expireRooms(time.Now())
	room, exists := rooms[id]
	if !exists || len(room.Clients) >= 2 {
		http.Error(w, "Room not found or expired", http.StatusNotFound)
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
	currRoom, alreadyExists := rooms[id]
	if !alreadyExists || (len(currRoom.Clients) == 0 && !time.Now().Before(currRoom.ExpiresAt)) {
		delete(rooms, id)
		conn.Close()
		return
	}

	if len(currRoom.Clients) >= 2 {
		fmt.Printf("Room %s is unavailable, closing connection\n", currRoom.Name)
		conn.Close()
		return
	}

	currRoom.Clients[conn] = true
	defer closeConnectionsInRoom(id)

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
