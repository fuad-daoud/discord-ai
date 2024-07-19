package commands

import (
	"context"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/youtube"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
	"time"
)

func play(call *cohere.CommandCall) {
	dlog.Log.Info("starting play function")

	toolCall := call.ToolCall

	conn := platform.Client().VoiceManager().GetConn(call.ExtraProperties.GuildId)
	if conn == nil {
		voiceState, b := platform.Cache().VoiceState(call.ExtraProperties.GuildId, call.ExtraProperties.UserId)
		if !b {
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "user is not in a voice channel, the user shall be in a voice channel first",
					},
				},
			}
			return
		}

		botState, botStateOk := platform.Cache().VoiceState(call.ExtraProperties.GuildId, platform.Client().ApplicationID())
		userState, userStateOk := platform.Cache().VoiceState(call.ExtraProperties.GuildId, call.ExtraProperties.UserId)

		if botStateOk && userStateOk && (botState.ChannelID.String() == userState.ChannelID.String()) {
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "already in the voice channel",
					},
				},
			}
			return
		}

		dlog.Log.Info("joining voice channel")

		conn = platform.Client().VoiceManager().CreateConn(call.ExtraProperties.GuildId)

		if err := conn.Open(context.Background(), *voiceState.ChannelID, false, false); err != nil {
			dlog.Log.Error("error opening voice channel", "error", err)
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "error connecting to voice channel",
					},
				},
			}
			return
		}
		if err := conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone); err != nil {
			dlog.Log.Error("error setting speaking flag", "error", err)
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "error setting speaking flag",
					},
				},
			}
			return
		}
		dlog.Log.Info("set speaking successfully")
		if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
			dlog.Log.Error("failed to write silence audio frame", "error", err)
			cohere.Result <- &cohere.CommandResult{
				Call: call.ToolCall,
				Outputs: []map[string]interface{}{
					{
						"Success":     false,
						"Description": "error sending silence",
					},
				},
			}
			return
		}
		dlog.Log.Info("wrote silent frame successfully")
	}

	packets := make([][]byte, 0)
	y := youtube.Youtube{
		Process: func(seg []byte) {
			packets = append(packets, seg)
		},
		Progress: func(percentage float64) {
			dlog.Log.Info("Progress ", "percentage", percentage)
			platform.Rest().CreateMessage(snowflake.MustParse("1252273230727876619"), discord.MessageCreate{
				Content: fmt.Sprintf("Downloading %v%%", percentage),
			})
		},
		ProgressError: func(input string) {
			dlog.Log.Error("something wrong happened", "input", input)
		},
	}

	y.Play("22tVWwmTie8")

	player := youtube.DefaultPlayer{
		Segments: &packets,
		Conn:     conn,
	}
	player.Run()
	time.Sleep(2 * time.Second)
	player.Pause()
	go platform.HandleDeepgramVoicePackets(conn, call.ExtraProperties.MessageId)

	cohere.Result <- &cohere.CommandResult{
		Call: toolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "song playing successfully",
			},
		},
	}
}
