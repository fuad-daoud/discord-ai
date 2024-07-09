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

	dlog.Info("Deepgram", "Link", mr.Channel.Alternatives[0].Confidence, "sentence", sentence)
	dlog.Info("Deepgram", "isFinal", mr.IsFinal)
	dlog.Info("Deepgram", "isSpeachFinal", mr.SpeechFinal)
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
	dlog.Info("[Metadata] Received")
	dlog.Info("Metadata", "RequestID", strings.TrimSpace(md.RequestID))
	dlog.Info("Metadata", "Channels", md.Channels)
	dlog.Info("Metadata", "Created", strings.TrimSpace(md.Created))
	return nil
}

func (c *MyCallback) UtteranceEnd(ur *api.UtteranceEndResponse) error {
	dlog.Info("[UtteranceEnd] Received")
	return nil
}

func (c *MyCallback) Error(er *api.ErrorResponse) error {
	dlog.Error("[Error] Received")
	dlog.Error("", "Type", er.Type)
	dlog.Error("", "Message", er.Message)
	dlog.Error("", "Description", er.Description)
	return nil
}
