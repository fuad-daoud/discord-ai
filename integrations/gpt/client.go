package gpt

import (
	"fmt"
	"github.com/fuad-daoud/discord-ai/integrations/custom_http"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	Action     = make(chan FunctionInput)
	Response   = make(chan FunctionOutput)
	httpClient *custom_http.Client
)

func getHttpClient() custom_http.Client {
	if httpClient == nil {
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json; charset=utf-8"
		headers["Authorization"] = os.ExpandEnv("Bearer $OPENAI_API_KEY")
		headers["Openai-Beta"] = "assistants=v2"
		var c custom_http.Client = &custom_http.DefaultClient{
			BaseURL: "https://api.openai.com",
			Client:  &http.Client{},
			Headers: headers,
		}
		httpClient = &c
		return c
	}
	return *httpClient
}

func Detect(message, messageId, userId, threadId string) int {
	response := SendMessageFullCycle("D:"+message, messageId, userId, threadId)
	atoi, err := strconv.Atoi(response)
	if err != nil {
		slog.Warn("was not able to parse response not int", "response", response)
		return 0
	}
	slog.Info("on detect response: ", "perct", atoi)
	return atoi
}

func SendMessageFullCycle(message, messageId, userId, threadId string) string {
	slog.Info("send message got", "msg", message, "messageId", messageId, "userId", userId, "threadId", threadId)
	sendMessage(message, threadId)
	run := runThread(messageId, userId, threadId)
	messages := getMessages(run.ThreadId)
	return messages.Data[0].Content[0].Text.Value
}
func getMessages(threadId string) Messages {
	client := getHttpClient()
	req := client.GetRequest(fmt.Sprintf("/v1/threads/%s/messages", threadId))

	var messages Messages
	client.DoJson(req, &messages)
	return messages
}

func CreateThread() Thread {
	client := getHttpClient()
	req := client.PostEmptyRequest("/v1/threads")

	var thread Thread

	client.DoJson(req, &thread)
	return thread
}
func SubmitToolOutputs(toolCallId string, output OutputTool) *Run {
	body := strings.NewReader(fmt.Sprintf(
		`{
    "tool_outputs": [
      {
        "tool_call_id": "%s",
        "output": "%s"
      }
    ]
  }`, toolCallId, output.Description))
	client := getHttpClient()
	req := client.PostRequest(fmt.Sprintf("/v1/threads/%s/runs/%s/submit_tool_outputs", output.ThreadId, output.Id), body)

	var run Run

	client.DoJson(req, &run)

	return &run
}

func sendMessage(message string, threadId string) {
	body := strings.NewReader(fmt.Sprintf(
		`{
      "role": "user",
      "content": "%s"
    }`, message))
	client := getHttpClient()
	req := client.PostRequest(fmt.Sprintf("/v1/threads/%s/messages", threadId), body)
	client.Do(req)
}

func runThread(messageId, userId, threadId string) Run {
	body := strings.NewReader(fmt.Sprintf(
		`{
    "assistant_id": "%s"
  }`, os.Getenv("GPT_ASSISTANT")))
	client := getHttpClient()
	req := client.PostRequest(fmt.Sprintf("/v1/threads/%s/runs", threadId), body)

	var run Run
	client.DoJson(req, &run)
	waitTillDone(run, messageId, userId)
	return run
}

func waitTillDone(run Run, messageId, userId string) {
	for {
		status := run.Status
		slog.Info("run status ", "status", status)
		if status == "completed" {
			break
		}
		//queued, in_progress, requires_action, cancelling, cancelled, failed, completed, incomplete, or expired
		if status == "requires_action" {
			toolCalls := run.RequiredAction.SubmitToolOutputs.ToolCalls[0]
			Action <- FunctionInput{
				Function:  toolCalls.Function,
				UserId:    userId,
				MessageId: messageId,
			}
			select {
			case response := <-Response:

				tool := OutputTool{
					FunctionOutput: response,
					Id:             run.Id,
					ThreadId:       run.ThreadId,
				}

				run := SubmitToolOutputs(toolCalls.Id, tool)
				status = run.Status
			}
		}
		if status == "cancelled" || status == "failed" || status == "expired" || status == "cancelling" {
			slog.Error("stats run is not valid can't complete", "status", status)
			break
		}

		time.Sleep(100 * time.Millisecond)
		run = *getRun(run.ThreadId, run.Id)
	}
}

func getRun(threadId string, runId string) *Run {
	client := getHttpClient()
	req := client.GetRequest(fmt.Sprintf("/v1/threads/%s/runs/%s", threadId, runId))
	var run Run
	client.DoJson(req, &run)
	return &run
}

func CancelRunsForThread(threadId string) {
	runs := listRuns(threadId)
	for _, run := range runs.Data {
		if run.Status == "queued" || run.Status == "in_progress" || run.Status == "requires_action" || run.Status == "incomplete" {
			cancelRun(threadId, run.Id)
		}
	}
}

func cancelRun(threadId string, runId string) {
	client := getHttpClient()
	request := client.PostEmptyRequest(fmt.Sprintf("/v1/threads/%s/runs/%s/cancel", threadId, runId))
	var run Run
	client.DoJson(request, &run)
	slog.Info("cancel run ", "threadId", threadId, "runId", runId, "status", run.Status)
}

func listRuns(threadId string) Runs {
	client := getHttpClient()

	req := client.GetRequest(fmt.Sprintf("/v1/threads/%s/runs", threadId))

	var runs Runs
	client.DoJson(req, &runs)
	return runs
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

type Runs struct {
	Data []Run
}

type Run struct {
	Id             string `json:"id"`
	Status         string `json:"status"`
	ThreadId       string `json:"thread_id"`
	RequiredAction struct {
		Type              string `json:"type"`
		SubmitToolOutputs struct {
			ToolCalls []struct {
				Id       string   `json:"id"`
				Type     string   `json:"type"`
				Function Function `json:"function"`
			} `json:"tool_calls"`
		} `json:"submit_tool_outputs"`
	} `json:"required_action"`
	MetaData MetaData `json:"meta_data"`
}

type MetaData struct {
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
}

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type FunctionInput struct {
	Function
	UserId    string
	MessageId string
}

type FunctionOutput struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
}

type OutputTool struct {
	FunctionOutput
	Id       string `json:"id"`
	ThreadId string `json:"run_id"`
}
