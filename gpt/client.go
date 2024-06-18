package gpt

import (
	"fmt"
	"github.com/bwmarrin/discordgo/examples/voice_receive/custom_http"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client interface {
	CreateThread() Thread
	SendMessage(message string)
	RunThread() Run
	GetMessages() Messages
	GetRun() *Run
	SubmitToolOutputs(toolCallId string, output OutputTool) *Run
	CheckDone(data MetaData)
	SendMessageFullCycle(message string, metaData MetaData) string
	GetThreadId() string
	GetChanRequiredAction() chan Run
}

const (
	AssistantId = "asst_X1g2Iqb5z4KHfaxmL3bBBP3I"
)

type DefaultClient struct {
	Thread         Thread
	Run            Run
	Client         custom_http.Client
	RequiredAction chan Run
}

func MakeClient() Client {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = os.ExpandEnv("Bearer $OPENAI_API_KEY")
	headers["Openai-Beta"] = "assistants=v2"
	var client custom_http.Client = &custom_http.DefaultClient{
		BaseURL: "https://api.openai.com",
		Client:  &http.Client{},
		Headers: headers,
	}

	return &DefaultClient{
		Thread:         Thread{},
		Run:            Run{},
		Client:         client,
		RequiredAction: make(chan Run),
	}
}

func (c *DefaultClient) GetChanRequiredAction() chan Run {
	return c.RequiredAction
}

func (c *DefaultClient) GetThreadId() string {
	return c.Thread.Id
}

func (c *DefaultClient) SendMessageFullCycle(message string, data MetaData) string {
	c.SendMessage(message)
	c.RunThread()
	c.CheckDone(data)
	messages := c.GetMessages()
	return messages.Data[0].Content[0].Text.Value
}
func (c *DefaultClient) GetMessages() Messages {
	req := c.Client.GetRequest(fmt.Sprintf("/v1/threads/%s/messages", c.Run.ThreadId))

	var messages Messages
	c.Client.DoJson(req, &messages)
	return messages
}

func (c *DefaultClient) CreateThread() Thread {
	req := c.Client.PostEmptyRequest("/v1/threads")

	var thread Thread

	c.Client.DoJson(req, &thread)
	c.Thread = thread
	return thread
}
func (c *DefaultClient) SubmitToolOutputs(toolCallId string, output OutputTool) *Run {
	body := strings.NewReader(fmt.Sprintf(
		`{
    "tool_outputs": [
      {
        "tool_call_id": "%s",
        "output": "%s"
      }
    ]
  }`, toolCallId, output.Description))
	req := c.Client.PostRequest(fmt.Sprintf("/v1/threads/%s/runs/%s/submit_tool_outputs", c.Run.ThreadId, c.Run.Id), body)

	var run Run

	c.Client.DoJson(req, &run)
	c.Run = run
	c.Thread.Id = run.ThreadId
	return &run
}

func (c *DefaultClient) SendMessage(message string) {
	body := strings.NewReader(fmt.Sprintf(
		`{
      "role": "user",
      "content": "%s"
    }`, message))
	req := c.Client.PostRequest(fmt.Sprintf("/v1/threads/%s/messages", c.Thread.Id), body)
	var not_used any
	c.Client.DoJson(req, &not_used)
}

func (c *DefaultClient) RunThread() Run {
	body := strings.NewReader(fmt.Sprintf(
		`{
    "assistant_id": "%s"
  }`, AssistantId))
	req := c.Client.PostRequest(fmt.Sprintf("/v1/threads/%s/runs", c.Thread.Id), body)

	var run Run
	c.Client.DoJson(req, &run)
	c.Run = run
	c.Thread.Id = run.ThreadId
	return run
}

func (c *DefaultClient) CheckDone(data MetaData) {
	for {
		status := c.Run.Status
		log.Println("run status ", status)
		if status == "completed" {
			break
		}
		//queued, in_progress, requires_action, cancelling, cancelled, failed, completed, incomplete, or expired
		if status == "requires_action" {
			c.Run.MetaData = data
			c.RequiredAction <- c.Run
			select {
			case newRun := <-c.RequiredAction:
				c.Run = newRun
				status = c.Run.Status
			}
		}
		if status == "cancelled" || status == "failed" || status == "expired" || status == "cancelling" {
			log.Fatal("stats run is not valid can't complete status:" + status)
		}

		time.Sleep(100 * time.Millisecond)
		c.GetRun()
	}
}

func (c *DefaultClient) GetRun() *Run {
	req := c.Client.GetRequest(fmt.Sprintf("/v1/threads/%s/runs/%s", c.Run.ThreadId, c.Run.Id))

	c.Client.DoJson(req, &c.Run)
	return &c.Run
}

type Thread struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

type Messages struct {
	Data []struct {
		Id      string `json:"id"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text struct {
				Value string `json:"value"`
			} `json:"text"`
		} `json:"content"`
	} `json:"data"`
}

type Run struct {
	Id             string `json:"id"`
	Status         string `json:"status"`
	ThreadId       string `json:"thread_id"`
	RequiredAction struct {
		Type              string `json:"type"`
		SubmitToolOutputs struct {
			ToolCalls []struct {
				Id       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name string `json:"name"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"submit_tool_outputs"`
	} `json:"required_action"`
	MetaData MetaData `json:"meta_data"`
}

type MetaData struct {
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
}

type OutputTool struct {
	Success     string `json:"success"`
	Description string `json:"description"`
}
