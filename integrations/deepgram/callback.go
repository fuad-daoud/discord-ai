package deepgram

import (
	api "github.com/deepgram/deepgram-go-sdk/pkg/api/live/v1/interfaces"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"log"
	"strings"
)

type MyCallback struct {
	Builder     strings.Builder
	SpeechFinal chan bool
	sentence    string
}

func (c *MyCallback) Message(mr *api.MessageResponse) error {
	sentence := strings.TrimSpace(mr.Channel.Alternatives[0].Transcript)
	if len(mr.Channel.Alternatives) == 0 || len(sentence) == 0 {
		return nil
	}

	dlog.Log.Info("Deepgram", "Link", mr.Channel.Alternatives[0].Confidence, "sentence", sentence)
	dlog.Log.Info("Deepgram", "isFinal", mr.IsFinal)
	dlog.Log.Info("Deepgram", "isSpeachFinal", mr.SpeechFinal)
	c.Builder.WriteString(sentence)
	c.sentence = sentence
	if mr.SpeechFinal {
		c.SpeechFinal <- mr.SpeechFinal
	}
	return nil
}

func (c *MyCallback) SpeechStarted(ssr *api.SpeechStartedResponse) error {
	log.Printf("Speech started")
	return nil
}

func (c *MyCallback) Close(cr *api.CloseResponse) error {
	log.Printf("Close called")
	return nil
}

func (c *MyCallback) Open(or *api.OpenResponse) error {
	log.Printf("Open called")
	return nil
}

func (c *MyCallback) UnhandledEvent(byData []byte) error {
	log.Printf("Unhandled event")
	return nil
}

func (c *MyCallback) Metadata(md *api.MetadataResponse) error {
	dlog.Log.Info("[Metadata] Received")
	dlog.Log.Info("Metadata", "RequestID", strings.TrimSpace(md.RequestID))
	dlog.Log.Info("Metadata", "Channels", md.Channels)
	dlog.Log.Info("Metadata", "Created", strings.TrimSpace(md.Created))
	return nil
}

func (c *MyCallback) UtteranceEnd(ur *api.UtteranceEndResponse) error {
	dlog.Log.Info("[UtteranceEnd] Received")
	return nil
}

func (c *MyCallback) Error(er *api.ErrorResponse) error {
	dlog.Log.Error("[Error] Received")
	dlog.Log.Error("", "Type", er.Type)
	dlog.Log.Error("", "Message", er.Message)
	dlog.Log.Error("", "Description", er.Description)
	return nil
}
