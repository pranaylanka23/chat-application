package main

import (
	"chatApplication/server"
	"log"
	"net/http"
)

func main() {
	// Start listening for broadcast messages in background
	go server.HandleMessages()

	http.HandleFunc("/ws", server.HandleWebSocket)

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
