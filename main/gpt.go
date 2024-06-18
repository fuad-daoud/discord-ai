package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo/examples/voice_receive/custom_http"
	"github.com/bwmarrin/discordgo/examples/voice_receive/gpt"
	"net/http"
	"os"
)

const (
	Text        = "How can you help me?"
	AssistantId = "asst_X1g2Iqb5z4KHfaxmL3bBBP3I"
)

// enum
// status = queued, in_progress, requires_action, cancelling, cancelled, failed, completed, incomplete,  expired

func main() {

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = os.ExpandEnv("Bearer $OPENAI_API_KEY")
	headers["Openai-Beta"] = "assistants=v2"
	var client custom_http.Client = &custom_http.DefaultClient{
		BaseURL: "https://api.openai.com",
		Client:  &http.Client{},
		Headers: headers,
	}

	var gptClient gpt.Client = &gpt.DefaultClient{
		Thread: gpt.Thread{},
		Run:    gpt.Run{},
		Client: client,
	}

	gptClient.CreateThread()

	gptClient.SendMessage("Do you know what day it is?")
	gptClient.RunThread()
	gptClient.CheckDone()
	messages := gptClient.GetMessages()
	fmt.Println(messages)
}
