package cohere

import (
	"bufio"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/disgoorg/snowflake/v2"
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
	MessageId snowflake.ID
	UserId    snowflake.ID
	GuildId   snowflake.ID
	ChannelId snowflake.ID
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
			Description: "if not provided use call command_search to get the link first, play a youtube video given its link",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{
				"information": {
					Description: cohere.String("information about the song or the youtube link of a video"),
					Type:        "str",
					Required:    cohere.Bool(true),
				},
			},
		},
		{
			Name:        "command_search",
			Description: "find youtube link",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{
				"information": {
					Description: cohere.String("information about the song or the youtube link of a video"),
					Type:        "str",
					Required:    cohere.Bool(true),
				},
			},
		},
		{
			Name:                 "command_pause",
			Description:          "pause the current playing song",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
		},
		{
			Name:                 "command_stop",
			Description:          "stop the current playing song",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
		},
		{
			Name:                 "command_resume",
			Description:          "resume the current playing song",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
		},
		{
			Name:                 "command_skip",
			Description:          "skip the current playing song",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
		},
		{
			Name:                 "command_queue",
			Description:          "list songs in queue, list them in each on separate line",
			ParameterDefinitions: map[string]*cohere.ToolParameterDefinitionsValue{},
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
		dlog.Log.Error(err.Error())
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
