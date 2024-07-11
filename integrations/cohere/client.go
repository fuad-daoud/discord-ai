package cohere

import (
	"bufio"
	"errors"
	"fmt"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/cohere-ai/cohere-go/v2/client"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"io"
	"os"
	"strings"
)

type CommandCall struct {
	*cohere.ToolCall
	ExtraProperties Properties
}

type Properties struct {
	MessageId string `json:"message_id"`
	UserId    string `json:"user_id"`
	GuildId   string `json:"guild_id"`
}
type StreamResult struct {
	Message  *strings.Builder
	Finished bool
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

func String(s string) *string {
	return cohere.String(s)
}

func Stream(message, conversationId string, prop Properties, start func(), end func()) *StreamResult {
	co := client.NewClient(client.WithToken("xLPWbInVLTliZHK8JbYxYrtoEpu6K4Y8KFjJVJZ5"))
	request := &cohere.ChatStreamRequest{
		Message:        message,
		Model:          cohere.String("command-r-plus"),
		Preamble:       cohere.String(readInst()),
		ConversationId: cohere.String(conversationId),
		Temperature:    cohere.Float64(0.99),
		Tools:          tools,
	}

	result := &StreamResult{
		Message: &strings.Builder{},
	}
	go processStream(result, request, prop, co, start, end)
	return result
}

func processStream(result *StreamResult, request *cohere.ChatStreamRequest, prop Properties, co *client.Client, start func(), end func()) {
	chatStream, err := co.ChatStream(context.Background(), request)
	if err != nil {
		dlog.Error(err.Error())
		panic(err)
	}

	for {
		if result.Finished {
			return
		}
		response, err := chatStream.Recv()
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
		dlog.Info("got event", "eventType", response.EventType, "response", response)
		go processStreamEvent(result, response, request, prop, co, start, end)
		if response.EventType == "stream-end" || response.EventType == "tool-calls-generation" {
			break
		}
	}
}

func processStreamEvent(result *StreamResult, response cohere.StreamedChatResponse, request *cohere.ChatStreamRequest, prop Properties, co *client.Client, start func(), end func()) {
	switch response.EventType {
	case "tool-calls-generation":
		{
			toolCalls := response.ToolCallsGeneration.ToolCalls
			request.ToolResults = make([]*cohere.ToolResult, len(toolCalls))
			for index, toolCall := range toolCalls {
				Call <- &CommandCall{
					ToolCall:        toolCall,
					ExtraProperties: prop,
				}
				request.ToolResults[index] = <-Result
			}
			request.Message = ""
			chatStream, err := co.ChatStream(context.TODO(), request)
			if err != nil {
				dlog.Error(err.Error())
				panic(err)
			}
			for {
				if result.Finished {
					return
				}
				response, err := chatStream.Recv()
				if err != nil && !errors.Is(err, io.EOF) {
					panic(err)
				}
				dlog.Info("got event", "eventType", response.EventType, "response", response)
				go processStreamEvent(result, response, request, prop, co, start, end)
				if response.EventType == "stream-end" {
					break
				}
			}
			break
		}
	case "stream-start":
		start()
		break
	case "stream-end":
		end()
		result.Finished = true
		break
	case "text-generation":
		result.Message.WriteString(response.TextGeneration.Text)
		break
	default:
		break
	}
}

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
		dlog.Error(err.Error())
		panic(err)
		return "Something wrong happened contact Admin"
	}
	dlog.Debug("Response from cohere", "resp", fmt.Sprintf("%+v", resp))
	if resp.ToolCalls != nil {
		request.ToolResults = make([]*cohere.ToolResult, len(resp.ToolCalls))
		for index, toolCall := range resp.ToolCalls {
			Call <- &CommandCall{
				ToolCall: toolCall,
				//ExtraProperties: map[string]interface{}{
				//	"messageId": messageId,
				//	"userId":    userId,
				//},
			}
			request.ToolResults[index] = <-Result
		}
		request.Message = ""
		resp, err = co.Chat(context.TODO(), request)
		if err != nil {
			dlog.Error(err.Error())
			panic(err)
			return "Something wrong happened contact Admin"
		}

	}
	return resp.Text
}

func Send(message, messageId, userId, conversationId string) string {
	return chat(message, messageId, userId, conversationId)
}

func readInst() string {
	readFile, err := os.Open(os.Getenv("COHERE_INST"))
	defer readFile.Close()
	if err != nil {
		dlog.Error(err.Error())
		panic(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	builder := strings.Builder{}
	for fileScanner.Scan() {
		builder.WriteString(fileScanner.Text())
	}

	return builder.String()
}
