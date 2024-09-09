package main

import (
	"flag"
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/remote"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

type clientMap map[string]*actor.PID
type userMap map[string]string
type chatroomMap map[string]*server

type server struct {
	clients    clientMap // key: address value: *pid
	users      userMap   // key: address value: username
	logger     *slog.Logger
	chatroomID string
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for this example
		},
	}
	chatrooms = make(chatroomMap)
)

func newServer(chatroomID string) actor.Receiver {
	return &server{
		clients:    make(clientMap),
		users:      make(userMap),
		logger:     slog.Default(),
		chatroomID: chatroomID,
	}
}

func (s *server) Receive(ctx *actor.Context) {
	// Implementing the actor message handling logic
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, chatroomID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Error reading message", "error", err)
			return
		}
		if messageType == websocket.TextMessage {
			slog.Info("Received message", "message", string(p), "chatroom", chatroomID)
		}
	}
}

func createChatroom(w http.ResponseWriter, r *http.Request) {
	// Extracting chatroom ID from path parameters
	chatroomID := r.URL.Path[len("/chatroom/"):]
	if _, err := uuid.Parse(chatroomID); err != nil {
		http.Error(w, "Invalid UUID format for chatroom ID", http.StatusBadRequest)
		return
	}

	// If chatroom does not exist, create a new one
	if _, exists := chatrooms[chatroomID]; !exists {
		chatrooms[chatroomID] = newServer(chatroomID).(*server)
		slog.Info("Created new chatroom", "chatroomID", chatroomID)
	}
}

func main() {
	var (
		listenAtWebsockets = flag.String("listen", "127.0.0.1:8001", "")
		listenAtHTTP       = flag.String("listen", "127.0.0.1:8000", "")
	)
	flag.Parse()

	rem := remote.New(*listenAtWebsockets, remote.NewConfig())
	e, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		panic(err)
	}

	// Setup routing using the new Go 1.22 features
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chatroom/{id}", createChatroom)
	mux.HandleFunc("GET /ws/chatroom/{id}", func(w http.ResponseWriter, r *http.Request) {
		chatroomID := r.URL.Path[len("/ws/chatroom/"):]
		if _, exists := chatrooms[chatroomID]; exists {
			handleWebSocket(w, r, chatroomID)
		} else {
			http.Error(w, "Chatroom not found", http.StatusNotFound)
		}
	})

	go func() {
		slog.Info("Starting HTTP server", "address", *listenAtHTTP)
		if err := http.ListenAndServe(*listenAtHTTP, mux); err != nil {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	select {}
}
