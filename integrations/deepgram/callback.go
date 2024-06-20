package deepgram

import (
	api "github.com/deepgram/deepgram-go-sdk/pkg/api/live/v1/interfaces"
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

	log.Printf("Deepgram %f: %s\n\n", mr.Channel.Alternatives[0].Confidence, sentence)
	log.Printf("Deepgram isFinal %v", mr.IsFinal)
	log.Printf("Deepgram isSpeachFinal %v", mr.SpeechFinal)
	c.Builder.WriteString(sentence)
	c.sentence = sentence
	if mr.SpeechFinal {
		c.SpeechFinal <- mr.SpeechFinal
	}
	// IsFinal => finished a sentence
	// SpeechFinal => finished this audio transcription
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
	log.Printf("\n[Metadata] Received\n")
	log.Printf("Metadata.RequestID: %s\n", strings.TrimSpace(md.RequestID))
	log.Printf("Metadata.Channels: %d\n", md.Channels)
	log.Printf("Metadata.Created: %s\n\n", strings.TrimSpace(md.Created))
	return nil
}

func (c *MyCallback) UtteranceEnd(ur *api.UtteranceEndResponse) error {
	log.Printf("\n[UtteranceEnd] Received\n")
	return nil
}

func (c *MyCallback) Error(er *api.ErrorResponse) error {
	log.Printf("\n[Error] Received\n")
	log.Printf("Error.Type: %s\n", er.Type)
	log.Printf("Error.Message: %s\n", er.Message)
	log.Printf("Error.Description: %s\n\n", er.Description)
	return nil
}
