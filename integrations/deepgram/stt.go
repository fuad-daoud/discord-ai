package deepgram

import (
	"errors"
	"github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	deepgramLive "github.com/deepgram/deepgram-go-sdk/pkg/client/live"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"os"
	"strings"
	"sync"
)

var (
	clients              = make(map[string]*deepgramLive.Client)
	silentPacketsCounter = make(map[string]int)
	silentPacket         = []byte{0xF8, 0xFF, 0xFE}
	m                    = sync.Mutex{}
)

const (
	silentPacketTime = 20
	silentSecond     = 50 * silentPacketTime
	silentHalfSecond = 25 * silentPacketTime
)

func Write(p []byte, userId string) {

	deepgramClient := clients[userId]

	_, err := deepgramClient.Write(p)
	if err != nil {
		if clients[userId] == nil {
			dlog.Log.Info("Stopping deepgram writing because deepgramLive is stopped")
		} else if strings.EqualFold(err.Error(), "websocket: close sent") {
			dlog.Log.Info("Stopping deepgram writing because", "err", err.Error())
		} else {
			return
		}
	}
	//dlog.Log.Info("deepgram reading", "bytes", voiceBytes)
}

func MakeClient(userId string, finishedCallback FinishedCallBack) (*deepgramLive.Client, error) {
	m.Lock()
	defer m.Unlock()
	client, ok := clients[userId]
	if !ok {
		// Configuration for the Client deepgramLive
		ctx := context.Background()

		apiKey := os.Getenv("DEEPGRAM_API_KEY")

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
			Endpointing:     "100",
			Extra:           nil,
			FillerWords:     true,
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

		callback := &MyCallback{Builder: strings.Builder{}, SpeechFinal: make(chan bool)}
		dgClient, err := deepgramLive.New(ctx, apiKey, &clientOptions, &transcriptOptions, callback)
		if err != nil {
			return nil, err
		}
		wsconn := dgClient.Connect()
		if wsconn == false {
			return nil, errors.New("deepgramLive.Connect failed")
		}

		dlog.Log.Info("Connected to deepgram client!", "userId", userId)

		go stopWhenFinished(userId, callback, finishedCallback)
		clients[userId] = dgClient
		return dgClient, nil
	}
	return client, nil
}

type FinishedCallBack func(message string, userId string)

func stopWhenFinished(userId string, callback *MyCallback, finishedCallback FinishedCallBack) {
	finished := <-callback.SpeechFinal
	dlog.Log.Info("channel SpeechFinal triggered", "speechfinal", finished)
	if finished {
		finishedCallback(callback.sentence, userId)
		StopUser(userId)
	}
}

func Stop() {
	for userId, _ := range clients {
		StopUser(userId)
		dlog.Log.Info("Stopped Client deepgram for", "userId", userId)
	}
	clients = make(map[string]*deepgramLive.Client)
}

func StopUser(userId string) {
	err := clients[userId].Finalize()
	if err != nil {
		dlog.Log.Error(err.Error())
	}
	//clients[userId].Stop()
	delete(clients, userId)
	dlog.Log.Info("removed client deepgram for", "userId", userId)
}
