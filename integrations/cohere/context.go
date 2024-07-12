package cohere

import (
	"bufio"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
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
	Message string
	Type    StreamResultType
}

type StreamResultType = string

const (
	Start StreamResultType = "start"
	End   StreamResultType = "end"
	Text  StreamResultType = "text"
)

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

type StreamContext struct {
	prop     Properties
	request  *cohere.ChatStreamRequest
	ctx      context.Context
	response cohere.StreamedChatResponse
	result   chan StreamResult
}

func (s *StreamContext) end(message string) {
	s.result <- StreamResult{
		Message: message,
		Type:    End,
	}
}
func (s *StreamContext) start() {
	s.result <- StreamResult{
		Type: Start,
	}
}
func (s *StreamContext) Text(message string) {
	s.result <- StreamResult{
		Type:    Text,
		Message: message,
	}
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

func String(s string) *string {
	return cohere.String(s)
}
