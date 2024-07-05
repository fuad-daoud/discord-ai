package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gage-technologies/mistral-go"
)

func main() {
	// If api key is empty it will load from MISTRAL_API_KEY env var
	client := mistral.NewMistralClientDefault(os.Getenv("MISTRAL_API_KEY"))

	log.Printf("Start Chat completion")

	// Example: Using Chat Completions
	var chatRes, err = client.Chat("open-mixtral-8x22b", []mistral.ChatMessage{{
		Role:    mistral.RoleSystem,
		Content: readInst(),
	}, {
		Content: "Luna, join",
		Role:    mistral.RoleUser,
	}, {
		Content: "Chat choises: Sure thing! I'm here to help!\n\nAre you sure you want me to join the voice channel you're currently in? Please confirm by saying \"Sure!\" or \"Yes!\" to proceed. If not, please let me know what you need help with instead.\n\n(Note: This response assumes that the user has already provided the necessary permissions for the bot to join the voice channel.)\n",
		Role:    mistral.RoleAssistant,
	},
		{
			Content: "Yes!",
			Role:    mistral.RoleUser,
		},
	}, &mistral.ChatRequestParams{
		Temperature: 0.7,
		TopP:        1,
		RandomSeed:  0,
		MaxTokens:   300,
		SafePrompt:  false,
		Tools: []mistral.Tool{
			{
				Type: "function",
				Function: mistral.Function{
					Name:        "join",
					Description: "join the voice channel the user is currently in",
					Parameters:  Parameters{},
				},
			},
			{
				Type: "function",
				Function: mistral.Function{
					Name:        "leave",
					Description: "leave voice channel",
					Parameters:  Parameters{},
				},
			}, {
				Type: "function",
				Function: mistral.Function{
					Name:        "play",
					Description: "play youtube video",
					Parameters: Parameters{
						Type: "object",
						Properties: Properties{
							Link: Property{
								Type:        "string",
								Description: "link to the youtube video",
							},
						},
						Required: []string{"link"},
					},
				},
			}},
		ToolChoice:     "auto",
		ResponseFormat: "",
	})
	if err != nil {
		log.Fatalf("Error getting chat completion: %v", err)
	}
	log.Printf("Chat completion: %+v\n", chatRes)

	log.Printf("Chat choises %+v\n", chatRes.Choices)
	log.Println()
	log.Println()
	log.Println()
	log.Printf("Chat choises: %+v\n", chatRes.Choices[0].Message.Content)
	log.Printf("Chat usage: %+v\n", chatRes.Usage)
	log.Printf("Chat toolcall: %+v\n", chatRes.Choices[0].Message.ToolCalls[0].Function.Arguments)

	// Example: Using Chat Completions Stream
	//chatResChan, err := client.ChatStream("mistral-tiny", []mistral.ChatMessage{{Content: "Hello, world!", Role: mistral.RoleUser}}, nil)
	//if err != nil {
	//	log.Fatalf("Error getting chat completion stream: %v", err)
	//}
	//
	//for chatResChunk := range chatResChan {
	//	if chatResChunk.Error != nil {
	//		log.Fatalf("Error while streaming response: %v", chatResChunk.Error)
	//	}
	//	log.Printf("Chat completion stream part: %+v\n", chatResChunk)
	//}

	// Example: Using Embeddings
	//embsRes, err := client.Embeddings("mistral-embed", []string{"Embed this sentence.", "As well as this one."})
	//if err != nil {
	//	log.Fatalf("Error getting embeddings: %v", err)
	//}

	//log.Printf("Embeddings response: %+v\n", embsRes)
}

type Parameters struct {
	Type       string
	Properties Properties
	Required   []string
}
type Properties struct {
	Link Property
	Sure Property
}

type Property struct {
	Type        string
	Description string
}

func readInst() string {
	readFile, err := os.Open("/home/fuad/GolandProjects/discord-ai/integrations/gpt/dev-inst.txt")

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	builder := strings.Builder{}
	for fileScanner.Scan() {
		builder.WriteString(fileScanner.Text())
	}

	readFile.Close()
	return builder.String()
}
