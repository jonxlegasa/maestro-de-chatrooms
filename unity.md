```go

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/examples/chat/types"
	"github.com/anthdm/hollywood/remote"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	switch msg := ctx.Message().(type) {
	case *types.Message:
		s.logger.Info("message received", "msg", msg.Msg, "from", ctx.Sender(), "chatroom", s.chatroomID)
		s.handleMessage(ctx)
	case *types.Disconnect:
		cAddr := ctx.Sender().GetAddress()
		pid, ok := s.clients[cAddr]
		if !ok {
			s.logger.Warn("unknown client disconnected", "client", pid.Address, "chatroom", s.chatroomID)
			return
		}
		username, ok := s.users[cAddr]
		if !ok {
			s.logger.Warn("unknown user disconnected", "client", pid.Address, "chatroom", s.chatroomID)
			return
		}
		s.logger.Info("client disconnected", "username", username, "chatroom", s.chatroomID)
		delete(s.clients, cAddr)
		delete(s.users, username)
	case *types.Connect:
		cAddr := ctx.Sender().GetAddress()
		if _, ok := s.clients[cAddr]; ok {
			s.logger.Warn("client already connected", "client", ctx.Sender().GetID(), "chatroom", s.chatroomID)
			return
		}
		if _, ok := s.users[cAddr]; ok {
			s.logger.Warn("user already connected", "client", ctx.Sender().GetID(), "chatroom", s.chatroomID)
			return
		}
		s.clients[cAddr] = ctx.Sender()
		s.users[cAddr] = msg.Username
		slog.Info("new client connected",
			"id", ctx.Sender().GetID(), "addr", ctx.Sender().GetAddress(), "sender", ctx.Sender(),
			"username", msg.Username, "chatroom", s.chatroomID,
		)
	}
}

func (s *server) handleMessage(ctx *actor.Context) {
	for _, pid := range s.clients {
		if !pid.Equals(ctx.Sender()) {
			s.logger.Info("forwarding message", "pid", pid.ID, "addr", pid.Address, "msg", ctx.Message(), "chatroom", s.chatroomID)
			ctx.Forward(pid)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, chatroomID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	// Here you would integrate the WebSocket connection with your actor system
	// For example, you could create a new actor for this connection and add it to the chatroom

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Error reading message", "error", err)
			return
		}
		if messageType == websocket.TextMessage {
			// Handle the message, possibly by sending it to the appropriate actor
			slog.Info("Received message", "message", string(p), "chatroom", chatroomID)
		}
	}
}

func main() {
	var (
		listenAt = flag.String("listen", "127.0.0.1:8000", "")
	)
	flag.Parse()

	rem := remote.New(*listenAt, remote.NewConfig())
	e, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/chatroom/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			http.Error(w, "Invalid chatroom ID", http.StatusBadRequest)
			return
		}
		chatroomID := parts[2]

		if _, err := uuid.Parse(chatroomID); err != nil {
			http.Error(w, "Invalid UUID format for chatroom ID", http.StatusBadRequest)
			return
		}

		if _, exists := chatrooms[chatroomID]; !exists {
			chatrooms[chatroomID] = newServer(chatroomID).(*server)
			e.Spawn(func() actor.Receiver { return chatrooms[chatroomID] }, fmt.Sprintf("server_%s", chatroomID))
		}

		handleWebSocket(w, r, chatroomID)
	})

	go func() {
		slog.Info("Starting HTTP server", "address", *listenAt)
		if err := http.ListenAndServe(*listenAt, nil); err != nil {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	select {}
}



```
```
```
## Mission
chatroom based on these new additons to the go lang language

Go 1.22 introduces several significant improvements to its HTTP handling, particularly within the `net/http` package, making it much easier to handle method-specific routing and path parameters.

1. **Method-Specific Routing**: You can now specify HTTP methods directly in route definitions. For example, you can define a handler like `mux.HandleFunc("GET /hello", handler)` that only accepts `GET` requests. If a different method is used, such as `POST`, the server will automatically return a `405 Method Not Allowed` response, simplifying code by removing the need to manually check methods inside handlers【6†source】【7†source】.

2. **Path Parameters**: The update introduces support for path parameters in routes using curly braces. For example, you can define a route like `/orders/{id}`, and the value of `id` can be extracted using the new `Request.PathValue` method. This eliminates the need for complex string manipulation to retrieve values from URLs【7†source】【10†source】.

3. **Wildcard Matching**: You can also use wildcards to match variable segments in the URL. For example, `{id}...` will match any number of segments, allowing for more flexible and dynamic routing scenarios【8†source】.

4. **Precedence in Routing**: The routing system now resolves conflicts using a "most specific wins" rule. This prioritizes routes that match more specific paths (like `/posts/latest` over `/posts/{id}`), ensuring predictable behavior when multiple routes could match a request【7†source】.

These updates make Go's `ServeMux` much more capable for handling modern web applications, without requiring third-party routers like Gorilla Mux for basic routing tasks.


## API Endpoints
based on these endpoints that will be websockets so that any client can connect to it and send messages:
/ws/chatroom/[id]
/ws/chatroom/[id]

then for these endpoints which are for creating chatrooms
POST /chatroom/[id] (creates an ID, Which you will spawn an actor with its own unique ID FROM that params0


