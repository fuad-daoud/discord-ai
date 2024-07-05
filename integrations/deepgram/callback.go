package deepgram

import (
	api "github.com/deepgram/deepgram-go-sdk/pkg/api/live/v1/interfaces"
	"log"
	"log/slog"
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

	slog.Info("Deepgram", "Link", mr.Channel.Alternatives[0].Confidence, "sentence", sentence)
	slog.Info("Deepgram", "isFinal", mr.IsFinal)
	slog.Info("Deepgram", "isSpeachFinal", mr.SpeechFinal)
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
	slog.Info("[Metadata] Received")
	slog.Info("Metadata", "RequestID", strings.TrimSpace(md.RequestID))
	slog.Info("Metadata", "Channels", md.Channels)
	slog.Info("Metadata", "Created", strings.TrimSpace(md.Created))
	return nil
}

func (c *MyCallback) UtteranceEnd(ur *api.UtteranceEndResponse) error {
	slog.Info("[UtteranceEnd] Received")
	return nil
}

func (c *MyCallback) Error(er *api.ErrorResponse) error {
	slog.Error("[Error] Received")
	slog.Error("", "Type", er.Type)
	slog.Error("", "Message", er.Message)
	slog.Error("", "Description", er.Description)
	return nil
}
