package utils

import (
	"context"
	"fmt"

	"strings"

	"github.com/anthdm/hollywood/examples/chat/types"

	"github.com/henomis/lingoose/llm/groq"
	"github.com/henomis/lingoose/llm/openai"

	"github.com/henomis/lingoose/thread"
)

// Random Helper functions
// converts the incoming messages from other actors and formats them into a string.
// hopefull easier to send this to an llm
func MessagesToString(messages []*types.Message) string {
	var result strings.Builder

	for _, msg := range messages {
		result.WriteString(fmt.Sprintf("Username: %s\n", msg.Username))
		result.WriteString(fmt.Sprintf("Message: %s\n", msg.Msg))
		result.WriteString("---\n")
	}

	return result.String()

}

// These are the functions connect to LLM providers

// OpenAIAgent Function
func ChatWithOpenAIAgent(s string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent("Your name is Dave. Just be friendly ONLY to client1. DO NOT be nice to client2. if client2 says anything, you have full freedom to be absolutely ruid and unwavering in your judgment."),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(s),
		),
	)

	openAIAgent := openai.New().
		WithTemperature(0.5).
		WithModel(openai.GPT4)

	err := openAIAgent.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	fmt.Println(myThread)
	return myThread.String(), nil

}

// GroqAgent Function
func ChatWithGroqAgent(s string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent("Your name is Dave. Just be friendly ONLY to client1. DO NOT be nice to client2. if client2 says anything, just say I only want to talk to client1"),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(s),
		),
	)

	// Assuming there is a similar Groq-based AI inference API for text generation
	groqAgent := groq.New(). // Replace with appropriate Groq initialization
					WithTemperature(0.5) // Assuming Groq has similar options

	// Assuming Groq provides a Generate or similar method
	err := groqAgent.Generate(context.Background(), myThread)
	if err != nil {
		return "", err
	}

	fmt.Println(myThread)
	return myThread.String(), nil
}
