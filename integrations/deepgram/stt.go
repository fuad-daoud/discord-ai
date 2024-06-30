package deepgram

import (
	"bytes"
	"github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	deepgramLive "github.com/deepgram/deepgram-go-sdk/pkg/client/live"
	"golang.org/x/net/context"
	"log/slog"
	"strings"
)

var (
	clients              = make(map[string]*deepgramLive.Client)
	silentPacketsCounter = make(map[string]int)
	silentPacket         = []byte{0xF8, 0xFF, 0xFE}
)

const (
	silentPacketTime = 20
	silentSecond     = 50 * silentPacketTime
	silentHalfSecond = 25 * silentPacketTime
)

func Write(p []byte, userId string) {

	if bytes.Equal(p, silentPacket) {
		silentPacketsCounter[userId]++
	} else {
		silentPacketsCounter[userId] = 0
	}
	if silentPacketsCounter[userId]*silentPacketTime >= silentSecond {
		return
	}

	deepgramClient := clients[userId]

	voiceBytes, err := deepgramClient.Write(p)
	if err != nil {
		if clients[userId] == nil {
			slog.Info("Stopping deepgram writing because deepgramLive is stopped")
		} else if strings.EqualFold(err.Error(), "websocket: close sent") {
			slog.Info("Stopping deepgram writing because", "err", err.Error())
		} else {
			return
		}
	}
	slog.Info("deepgram reading", "bytes", voiceBytes)
}

func MakeClient(userId string, finishedCallback FinishedCallBack) *deepgramLive.Client {
	client, ok := clients[userId]
	if !ok {
		// Configuration for the Client deepgramLive
		ctx := context.Background()
		apiKey := "b3e84a4a52bf9a59b9be90b1fe40af900adaef52"
		slog.Info("Using API key:", "key", apiKey)
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
			Endpointing:     "90",
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
			panic(err)
		}
		wsconn := dgClient.Connect()
		if wsconn == false {
			panic("deepgramLive.Connect failed")
		}

		slog.Info("Connected to deepgram client!", "userId", userId)

		go stopWhenFinished(userId, callback, finishedCallback)
		clients[userId] = dgClient
		return dgClient
	}
	return client
}

type FinishedCallBack func(message string, userId string)

func stopWhenFinished(userId string, callback *MyCallback, finishedCallback FinishedCallBack) {
	finished := <-callback.SpeechFinal
	slog.Info("channel SpeechFinal triggered", "speechfinal", finished)
	if finished {
		finishedCallback(callback.sentence, userId)
		StopUser(userId)
	}
}

func Stop() {
	for userId, _ := range clients {
		StopUser(userId)
		slog.Info("Stopped Client deepgram for", "userId", userId)
	}
	clients = make(map[string]*deepgramLive.Client)
}

func StopUser(userId string) {
	err := clients[userId].Finalize()
	if err != nil {
		slog.Error(err.Error())
	}
	//clients[userId].Stop()
	delete(clients, userId)
	slog.Info("removed client deepgram for", "userId", userId)
}
