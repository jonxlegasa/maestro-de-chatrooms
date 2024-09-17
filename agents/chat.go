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

type prompt struct {
	systemprompt     string
	chatroommessages string // For now I am just going to stream all the messages
	// to the LLM when the server broadcasts it back
}

type client struct {
	username  string
	serverPID *actor.PID
	logger    *slog.Logger
}

var receivedmessages []*types.Message

func newClient(username string, serverPID *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return &client{
			username:  username,
			serverPID: serverPID,
			logger:    slog.Default(),
		}
	}
}

func (c *client) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case *types.Message:
		fmt.Printf("%s: %s\n", msg.Username, msg.Msg)

		receivedmessages = append(receivedmessages, msg)

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
		input     = flag.String("input", "", "input text for the client")
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

	var (
		serverPID = actor.NewPID(*connectTo, "server/primary")
		// Spawn our client receiver
		clientPID = e.Spawn(newClient(*username, serverPID), "client", actor.WithID(*username))
	)

	receivedmessagesstr := utils.MessagesToString(receivedmessages)

	p := &prompt{
		systemprompt:     *input,
		chatroommessages: receivedmessagesstr,
	}

	fmt.Println("Number of received messages:", len(receivedmessages))

	//fmt.Println(receivedmessagesstr)
	aiResponse, err := utils.ChatWithGroqAgent(p.systemprompt, receivedmessagesstr)

	if err != nil {
		slog.Error("AI processing failed", "err", err)
		return
	}

	fmt.Println("AI Response generated:", aiResponse)

	// Create and send the response message

	aimessage := &types.Message{
		Msg:      aiResponse,
		Username: *username,
	}

	e.SendWithSender(serverPID, aimessage, clientPID)

	select {}
}
