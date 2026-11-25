package main

import "fmt"

func main() {
	fmt.Println("Running...")
	go startRoomExpiration()
	if err := runServer(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
