package utils

import (
	"context"
	"fmt"

	"os"

	"log"
	"strings"

	"github.com/anthdm/hollywood/examples/chat/types"

	"github.com/henomis/lingoose/llm/groq"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

// These are the functions connect to LLM providers

// OpenAIAgent Function
func ChatWithOpenAIAgent(sysprompt string, incomingprompt string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(sysprompt),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(incomingprompt),
		),
	)

	openaiagent := openai.New().
		WithTemperature(0.5).
		WithModel(openai.GPT4o)

	err := openaiagent.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	return myThread.String(), nil
}

// Gemini Function
func ChatWithAnthropicAgent(sysprompt string, incomingprompt string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(sysprompt),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(incomingprompt),
		),
	)

	// Assuming there is a similar Groq-based AI inference API for text generation
	anthropicagent := groq.New().WithModel("claude-3-5-sonnet-20240620"). // Replace with appropriate Groq initialization
										WithTemperature(0.5) // Assuming Groq has similar options

	err := anthropicagent.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	return myThread.String(), nil
}

// GroqAgent Function
func ChatWithGroqAgent(sysprompt string, incomingprompt string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(sysprompt),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(incomingprompt),
		),
	)

	// Assuming there is a similar Groq-based AI inference API for text generation
	groqagent := groq.New().WithModel("llama-3.1-8b-instant"). // Replace with appropriate Groq initialization
									WithTemperature(0.5) // Assuming Groq has similar options

	// Assuming Groq provides a Generate or similar method
	err := groqagent.Generate(context.Background(), myThread)
	if err != nil {
		return "", err
	}

	return myThread.String(), nil
}

// OpenAIAgent Function
func ChatWithGeminiAgent(sysprompt string, incomingprompt string) (string, error) {
	myThread := thread.New().AddMessage(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(sysprompt),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(incomingprompt),
		),
	)

	// Assuming there is a similar Groq-based AI inference API for text generation
	geminiagent := groq.New().WithModel("llama-3.1-8b-instant"). // Replace with appropriate Groq initialization
									WithTemperature(0.5) // Assuming Groq has similar options

	err := geminiagent.Generate(context.Background(), myThread)
	if err != nil {
		panic(err)
	}

	return myThread.String(), nil
}

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

func AppendMessagesToPrompt(message string, section string, filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	text := string(content)

	if strings.Contains(text, section) {
		parts := strings.Split(text, section)

		updatedsection := parts[1] + "\n" + message
		updatedtext := parts[0] + section + updatedsection
		err = os.WriteFile(filepath, []byte(updatedtext), 0644)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Message appended successfully!")
	} else {
		fmt.Println("Section not found!")
	}

}
