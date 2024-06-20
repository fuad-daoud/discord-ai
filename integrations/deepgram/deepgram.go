package deepgram

import (
	"github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/pkg/client/live"
	"golang.org/x/net/context"
	"log"
	"strings"
)

type Client interface {
	Write(p []byte, SSRC uint32, finishedCallback FinishedCallBack)
	StopSSRC(SSRC uint32)
	Stop()
	MapSSRC(SSRC int, userId string)
}
type defaultClient struct {
	clients map[uint32]*client.Client
	users   map[int]string
}

func MakeDefault() Client {
	return &defaultClient{}
}

func (dg *defaultClient) MapSSRC(SSRC int, userId string) {
	if dg.users == nil {
		dg.users = make(map[int]string)
	}
	dg.users[SSRC] = userId
}

func (dg *defaultClient) Write(p []byte, SSRC uint32, finishedCallback FinishedCallBack) {

	if dg.clients == nil {
		dg.clients = make(map[uint32]*client.Client)
	}
	deepgramClient := getDeepgramClient(dg, SSRC, finishedCallback)

	_, err := deepgramClient.Write(p)
	if err != nil {
		if dg.clients[SSRC] == nil {
			log.Println("Stopping deepgram writing because client is stopped")
		} else if strings.EqualFold(err.Error(), "websocket: close sent") {
			log.Println("Stopping deepgram writing because ", err.Error())
		} else {
			return
		}
	}
	//log.Printf("Client: %d bytes from deepgramClient \n", bytes)
}

func getDeepgramClient(dg *defaultClient, SSRC uint32, finishedCallback FinishedCallBack) *client.Client {
	if dg.clients[SSRC] == nil {
		// Configuration for the Client client
		ctx := context.Background()
		apiKey := "b3e84a4a52bf9a59b9be90b1fe40af900adaef52"
		log.Println("Using API key:", apiKey)
		clientOptions := interfaces.ClientOptions{
			APIKey:          "",
			Host:            "",
			APIVersion:      "",
			Path:            "",
			SkipServerAuth:  false,
			RedirectService: false,
			EnableKeepAlive: true,
		}
		transcriptOptions := interfaces.LiveTranscriptionOptions{
			Alternatives:    0,
			Callback:        "",
			CallbackMethod:  "",
			Channels:        0,
			Diarize:         false,
			DiarizeVersion:  "",
			Encoding:        "opus",
			Endpointing:     "80",
			Extra:           nil,
			FillerWords:     false,
			InterimResults:  true,
			Keywords:        nil,
			Language:        "en-US",
			Model:           "nova-2",
			Multichannel:    false,
			NoDelay:         false,
			Numerals:        false,
			ProfanityFilter: false,
			Punctuate:       false,
			Redact:          nil,
			Replace:         nil,
			SampleRate:      48000,
			Search:          nil,
			SmartFormat:     true,
			Tag:             nil,
			Tier:            "",
			UtteranceEndMs:  "",
			VadEvents:       false,
			Version:         "",
		}

		builder := strings.Builder{}
		callback := &MyCallback{Builder: builder, SpeechFinal: make(chan bool)}
		dgClient, err := client.New(ctx, apiKey, &clientOptions, &transcriptOptions, callback)
		if err != nil {
			panic(err)
		}
		wsconn := dgClient.Connect()
		if wsconn == false {
			panic("client.Connect failed")
		}

		log.Println("Connected!")

		dg.clients[SSRC] = dgClient

		go stopWhenFinished(dg, SSRC, callback, finishedCallback)

		return dgClient
	}
	return dg.clients[SSRC]
}

type FinishedCallBack func(message string, SRRC uint32)

func stopWhenFinished(dg *defaultClient, SSRC uint32, callback *MyCallback, finishedCallback FinishedCallBack) {
	finished := <-callback.SpeechFinal
	if finished {
		dg.StopSSRC(SSRC)

		finishedCallback(callback.sentence, SSRC)
	}

}

func (dg *defaultClient) Stop() {
	for SSRC, deepgramClient := range dg.clients {
		deepgramClient.Stop()
		log.Println("Stopped Client client for SSRC:", SSRC)
	}
	dg.clients = nil
}

func (dg *defaultClient) StopSSRC(SSRC uint32) {
	dg.clients[SSRC].Stop()
	log.Println("Stopped Client client for SSRC:", SSRC)
	delete(dg.clients, SSRC)
}
