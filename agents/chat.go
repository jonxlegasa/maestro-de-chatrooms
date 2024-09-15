package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/examples/chat/types"
	"github.com/anthdm/hollywood/remote"
	"github.com/jonxlegasa/maestro-de-chatrooms/utils"
)

type client struct {
	username         string
	serverPID        *actor.PID
	logger           *slog.Logger
	receivedMessages []*types.Message
}

func newClient(username string, serverPID *actor.PID) (actor.Producer, *client) {
	c := &client{
		username:  username,
		serverPID: serverPID,
		logger:    slog.Default(),
	}

	return func() actor.Receiver {
		return c
	}, c
}

func (c *client) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case *types.Message:
		// Lock the mutex before modifying receivedMessages
		c.receivedMessages = append(c.receivedMessages, msg)

		// Convert messages to a single string
		receivedMessagesStr := utils.MessagesToString(c.receivedMessages)

		// Process the messages with AI
		aiResponse, err := utils.ChatWithOpenAIAgent(receivedMessagesStr)
		if err != nil {
			c.logger.Error("AI processing failed", "err", err)
			return
		}

		// Create the response message
		responseMsg := &types.Message{
			Msg:      aiResponse,
			Username: c.username,
		}

		// Send the AI response back to the server
		ctx.Send(c.serverPID, responseMsg)

	case actor.Started:
		// Notify server that client has connected
		ctx.Send(c.serverPID, &types.Connect{
			Username: c.username,
		})
	case actor.Stopped:
		c.logger.Info("client stopped")
	}
}

func main() {
	var (
		listenAt  = flag.String("listen", "", "specify address to listen to, will pick a random port if not specified")
		connectTo = flag.String("connect", "127.0.0.1:4000", "the address of the server to connect to")
		username  = flag.String("username", os.Getenv("USER"), "")
	)
	flag.Parse()
	if *listenAt == "" {
		*listenAt = fmt.Sprintf("127.0.0.1:%d", rand.Int31n(50000)+10000)
	}
	rem := remote.New(*listenAt, remote.NewConfig())
	e, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		slog.Error("failed to create engine", "err", err)
		os.Exit(1)
	}

	// The process ID of the server
	serverPID := actor.NewPID(*connectTo, "server/primary")

	// Spawn the client actor
	clientActor, _ := newClient(*username, serverPID)
	e.Spawn(clientActor, "client", actor.WithID(*username))

	// Keep the main function running
	select {} // Block forever
}
