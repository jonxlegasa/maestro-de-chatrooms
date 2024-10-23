package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

		fmt.Println(string(message))

		if err != nil {
			log.Println("Error reading message:", err)
			delete(clients, conn)
			break
		}

		for client := range clients {
			fmt.Println(websocket.TextMessage)
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error broadcasting message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// ip := "192.168.111.14"
	ip := "127.0.0.1"
	port := "8000"

	fmt.Println("Server starting at " + ip + ":" + port)

	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(ip+":"+port, nil))
}
