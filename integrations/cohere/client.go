package cohere

import (
	"errors"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/cohere-ai/cohere-go/v2/client"
	"github.com/cohere-ai/cohere-go/v2/core"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"io"
)

func clientChatStream(ctx context.Context, request *cohere.ChatStreamRequest) *core.Stream[cohere.StreamedChatResponse] {
	co := client.NewClient(client.WithToken("xLPWbInVLTliZHK8JbYxYrtoEpu6K4Y8KFjJVJZ5"))
	chatStream, err := co.ChatStream(ctx, request)
	if err != nil {
		dlog.Error(err.Error())
		panic(err)
	}
	return chatStream
}

func StreamChat(message, conversationId string, prop Properties) chan StreamResult {
	request := &cohere.ChatStreamRequest{
		Message:        message,
		Model:          cohere.String("command-r-plus"),
		Preamble:       cohere.String(readInst()),
		ConversationId: cohere.String(conversationId),
		Temperature:    cohere.Float64(0.99),
		Tools:          tools,
	}

	results := make(chan StreamResult)
	go stream(&StreamContext{
		prop:    prop,
		request: request,
		ctx:     context.Background(),
		result:  results,
	})
	return results
}

func stream(context *StreamContext) {
	chatStream := clientChatStream(context.ctx, context.request)
	for {
		response, err := chatStream.Recv()
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
		if response.EventType == "stream-start" {
			dlog.Info("got event", "eventType", response.EventType, "response", response)
		}
		//dlog.Info("got event", "eventType", response.EventType, "response", response)
		context.response = response
		go handleStreamEvent(context)
		if response.EventType == "stream-end" || response.EventType == "tool-calls-generation" {
			dlog.Info("got event", "eventType", response.EventType, "response", response)
			break
		}
	}
}

func handleStreamEvent(context *StreamContext) {
	switch context.response.EventType {
	//case "tool-calls-generation":
	//	{
	//		toolCalls := context.response.ToolCallsGeneration.ToolCalls
	//		context.request.ToolResults = make([]*cohere.ToolResult, len(toolCalls))
	//		for index, toolCall := range toolCalls {
	//			Call <- &CommandCall{
	//				ToolCall:        toolCall,
	//				ExtraProperties: context.prop,
	//			}
	//			context.request.ToolResults[index] = <-Result
	//		}
	//		context.request.Message = ""
	//		chatStream := clientChatStream(context.ctx, context.request)
	//		for {
	//			response, err := chatStream.Recv()
	//			if err != nil && !errors.Is(err, io.EOF) {
	//				panic(err)
	//			}
	//			dlog.Debug("got event", "eventType", response.EventType, "response", response)
	//			context.response = response
	//			go handleStreamEvent(context)
	//			if response.EventType == "stream-end" {
	//				break
	//			}
	//		}
	//		break
	//	}
	case "stream-start":
		//dlog.Info("TEST started stream")
		go context.start()
		break
	case "stream-end":
		//dlog.Info("TEST stopped stream")
		go context.end(context.response.StreamEnd.Response.Text)
		break
	case "text-generation":
		go context.Text(context.response.TextGeneration.Text)
		break
	default:
		break
	}
}

func Send(_, _, _, _ string) string {
	//TODO: this is only used for when invoice try to use streaming instead when elvenlabs refill the quota
	return ""
}

//14:42:24.12247
//14:42:37.12377

//14:45:24.12247
//14:45:31.12317
