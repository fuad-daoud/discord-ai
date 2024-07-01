package platform

import (
	"bytes"
	"context"
	"errors"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pion/opus/pkg/oggreader"
	"io"
	"log/slog"
	"sync"
	"time"
)

var mutex = sync.Mutex{}

func Talk(conn voice.Conn, voiceReader io.ReadCloser, def func() error, undef func() error) error {
	sound, err := LoadSound(voiceReader)
	if err != nil {
		return err
	}
	err = PlaySound(conn, sound, def, undef)
	return err
}

func PlaySound(conn voice.Conn, buffer [][]byte, def func() error, undef func() error) (err error) {
	mutex.Lock()
	err = def()
	if err != nil {
		return err
	}
	err = conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone)
	if err != nil {
		return err
	}
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		return err
	}
	slog.Info("Starting writing packets")
	for _, buff := range buffer {
		_, err := conn.UDP().Write(buff)
		if err != nil {
			panic(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	slog.Info("Finished writing packets")
	mutex.Unlock()
	err = undef()
	if err != nil {
		return err
	}
	return nil
}

func LoadSound(voiceReader io.ReadCloser) ([][]byte, error) {
	ogg, _, err := oggreader.NewWith(voiceReader)
	if err != nil {
		return nil, err
	}
	var allSegments = make([][]byte, 0)
	for {
		segments, _, err := ogg.ParseNextPage()
		if errors.Is(err, io.EOF) {
			break
		} else if bytes.HasPrefix(segments[0], []byte("OpusTags")) {
			continue
		}

		for i := range segments {
			allSegments = append(allSegments, segments[i])
		}
	}
	return allSegments, nil
}

func deafen(guildId *snowflake.ID, channelId *snowflake.ID) func() error {
	return func() error {
		return Client().UpdateVoiceState(context.Background(), *guildId, channelId, false, true)
	}
}
func unDeafen(guildId *snowflake.ID, channelId *snowflake.ID) func() error {
	return func() error {
		return Client().UpdateVoiceState(context.Background(), *guildId, channelId, false, true)
	}
}

type updateDeafen struct {
}

type UpdateVoiceState func(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, selfMute bool, selfDeaf bool) error
