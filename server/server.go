package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Username string
}

var (
	clients   = make(map[*websocket.Conn]*Client)
	broadcast = make(chan Message)
	mutex     sync.Mutex
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte("Enter your username:"))
	_, usernameMsg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Username read error:", err)
		conn.Close()
		return
	}

	client := &Client{Conn: conn, Username: string(usernameMsg)}

	mutex.Lock()
	clients[conn] = client
	mutex.Unlock()

	broadcast <- Message{Sender: "System", Content: fmt.Sprintf("%s joined the chat", client.Username)}

	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		broadcast <- Message{Sender: "System", Content: fmt.Sprintf("%s left the chat", client.Username)}
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		broadcast <- Message{Sender: client.Username, Content: string(msg)}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		messageJSON, _ := json.Marshal(msg)
		mutex.Lock()
		for conn := range clients {
			err := conn.WriteMessage(websocket.TextMessage, messageJSON)
			if err != nil {
				log.Println("Write error:", err)
				conn.Close()
				delete(clients, conn)
			}
		}
		mutex.Unlock()
	}
}
