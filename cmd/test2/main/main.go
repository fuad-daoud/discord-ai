package main

import (
	"bufio"
	"context"
	"fmt"
	cohere "github.com/cohere-ai/cohere-go/v2"
	client "github.com/cohere-ai/cohere-go/v2/client"
	"log"
	"os"
	"strings"
)

func main() {
	// use X-Client-Name: The name of the project that is making the request.
	co := client.NewClient(client.WithToken("xLPWbInVLTliZHK8JbYxYrtoEpu6K4Y8KFjJVJZ5"))
	web(co)
	//bot(co)

}

func bot(co *client.Client) {
	conversationId := "10"
	model := "command-r-plus"
	instructions := readInst()
	temperature := 0.5
	//message := "join me luna"
	message := ""
	//tokenize, err := co.Tokenize(context.TODO(), &cohere.TokenizeRequest{
	//	Text:  message,
	//	Model: model,
	//})
	//if err != nil {
	//	return
	//}
	//log.Printf("tokenize: %+v", tokenize)
	fmt.Println()
	toolCall := cohere.ToolCall{
		Name:       "join",
		Parameters: nil,
	}
	resp, err := co.Chat(
		context.TODO(),
		&cohere.ChatRequest{
			Message:        message,
			Model:          &model,
			Preamble:       &instructions,
			ConversationId: &conversationId,
			Temperature:    &temperature,
			Tools: []*cohere.Tool{
				{
					Name:                 "command_join",
					Description:          "Join the voice channel the user is in",
					ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
				},
				{
					Name:                 "command_leave",
					Description:          "disconnect from the voice channel",
					ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
				},

				{
					Name:        "command_play",
					Description: "search and play a song off youtube",
					ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{
						"song": {
							Description: cohere.String("description of the song"),
							Type:        "str",
							Required:    cohere.Bool(true),
						},
					},
				},
			},
			ToolResults: []*cohere.ToolResult{
				{
					Call: &toolCall,
					Outputs: []map[string]interface{}{
						{
							"success":     "true",
							"description": "",
						},
					},
				},
			},
			ForceSingleStep: nil,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", resp)
}

func web(co *client.Client) {
	model := "command-r"
	temperature := 0.3
	message := "url of not afraid by eminem"
	tokenize, err := co.Tokenize(context.TODO(), &cohere.TokenizeRequest{
		Text:  message,
		Model: model,
	})
	if err != nil {
		return
	}
	log.Printf("tokenize: %+v", tokenize)
	fmt.Println()
	resp, err := co.Chat(
		context.TODO(),
		&cohere.ChatRequest{
			Message:          message,
			Model:            &model,
			Preamble:         nil,
			ChatHistory:      nil,
			ConversationId:   nil,
			PromptTruncation: nil,
			Connectors: []*cohere.ChatConnector{{
				Id: "web-search",
				Options: map[string]interface{}{
					"site": "youtube.com",
				},
			}},
			SearchQueriesOnly: cohere.Bool(false),
			Documents:         nil,
			CitationQuality:   cohere.ChatRequestCitationQualityAccurate.Ptr(),
			Temperature:       &temperature,
			MaxTokens:         nil,
			MaxInputTokens:    nil,
			K:                 nil,
			P:                 nil,
			Seed:              nil,
			StopSequences:     nil,
			FrequencyPenalty:  cohere.Float64(0),
			PresencePenalty:   nil,
			RawPrompting:      nil,
			ReturnPrompt:      nil,
			Tools:             nil,
			ToolResults:       nil,
			ForceSingleStep:   nil,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", resp)
}

func readInst() string {
	readFile, err := os.Open("/home/fuad/GolandProjects/discord-ai/integrations/gpt/cohere-inst.txt")

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

// input 188 77,598
// output 30 2,190
