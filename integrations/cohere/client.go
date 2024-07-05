package cohere

import (
	"bufio"
	"fmt"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/cohere-ai/cohere-go/v2/client"
	"golang.org/x/net/context"
	"log"
	"log/slog"
	"os"
	"strings"
)

type CommandCall struct {
	*cohere.ToolCall
	Properties map[string]interface{}
}
type CommandResult = cohere.ToolResult

var (
	Call   = make(chan *CommandCall)
	Result = make(chan *CommandResult)
)

var (
	tools = []*cohere.Tool{
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
			Description: "play a youtube video given information about it",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{
				"information": {
					Description: cohere.String("information about the song or the youtube link of a video"),
					Type:        "str",
					Required:    cohere.Bool(true),
				},
			},
		},
	}
)

func chat(message, messageId, userId, conversationId string) string {
	co := client.NewClient(client.WithToken("xLPWbInVLTliZHK8JbYxYrtoEpu6K4Y8KFjJVJZ5"))

	request := &cohere.ChatRequest{
		Message:        message,
		Model:          cohere.String("command-r-plus"),
		Preamble:       cohere.String(readInst()),
		ConversationId: cohere.String(conversationId),
		Temperature:    cohere.Float64(0.5),
		Tools:          tools,
	}
	resp, err := co.Chat(context.TODO(), request)
	if err != nil {
		log.Fatal(err)
		return "Something wrong happened contact Admin"
	}
	slog.Debug("Response from cohere", "resp", fmt.Sprintf("%+v", resp))
	if resp.ToolCalls != nil {
		request.ToolResults = make([]*cohere.ToolResult, len(resp.ToolCalls))
		for index, toolCall := range resp.ToolCalls {
			Call <- &CommandCall{
				ToolCall: toolCall,
				Properties: map[string]interface{}{
					"messageId": messageId,
					"userId":    userId,
				},
			}
			request.ToolResults[index] = <-Result
		}
		request.Message = ""
		resp, err = co.Chat(context.TODO(), request)
		if err != nil {
			log.Fatal(err)
			return "Something wrong happened contact Admin"
		}

	}
	return resp.Text
}

func Send(message, messageId, userId, conversationId string) string {
	return chat(message, messageId, userId, conversationId)
}

func readInst() string {
	readFile, err := os.Open("/home/fuad/GolandProjects/discord-ai/integrations/cohere/cohere-inst.txt")

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
