package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			delete(clients, conn)
			break
		}

		// Broadcast message to all clients
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error broadcasting message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {

	ip := "192.168.111.14"
	port := "8000"
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(ip+":"+port, nil))
}
