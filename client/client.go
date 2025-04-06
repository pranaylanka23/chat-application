package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

func main() {
	serverURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		os.Exit(0)
	}()

	_, prompt, _ := conn.ReadMessage()
	fmt.Print(string(prompt) + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text()))

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			var m Message
			json.Unmarshal(msg, &m)
			fmt.Printf("%s: %s\n", m.Sender, m.Content)
		}
	}()

	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" {
			log.Println("Exiting chat...")
			break
		}
		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
