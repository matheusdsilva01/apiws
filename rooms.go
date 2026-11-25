package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const roomTTL = 8 * time.Minute
const roomCleanupInterval = 30 * time.Second

type Room struct {
	Name      string
	Clients   map[*websocket.Conn]bool
	ExpiresAt time.Time
}

var rooms = make(map[string]Room)

func expireRooms(now time.Time) {
	for id, room := range rooms {
		if len(room.Clients) == 0 && !now.Before(room.ExpiresAt) {
			delete(rooms, id)
		}
	}
}

func startRoomExpiration() {
	ticker := time.NewTicker(roomCleanupInterval)
	defer ticker.Stop()

	for now := range ticker.C {
		expireRooms(now)
	}
}
