package commands

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/youtube"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
	"io"
	"os"
)

func play(call *cohere.CommandCall) {
	dlog.Log.Info("starting play function")
	dlog.Log.Info("play call", "params", call.ToolCall.Parameters)
	toolCall := call.ToolCall

	//message, err := platform.Rest().CreateMessage(call.ExtraProperties.ChannelId, discord.MessageCreate{
	//	Content: "Downloading ...",
	//})
	//if err != nil {
	//	panic(err)
	//}
	packets := make([][]byte, 0)
	open, err := os.Open("test.opus")
	if err != nil {
		panic(err)
	}
	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(open, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := open.Close()
			if err != nil {
				panic(err)
			}
			break
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			panic(err)
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(open, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			panic(err)
		}

		// Append encoded pcm data to the buffer.
		packets = append(packets, InBuf)
	}

	//y := youtube.Youtube{
	//	Process: func(seg []byte) {
	//		packets = append(packets, seg)
	//		go func() {
	//			if err != nil {
	//				dlog.Log.Error("faced an err writing", "err", err)
	//			}
	//		}()
	//	},
	//	Progress: func(percentage float64) {
	//		dlog.Log.Info("Progress ", "percentage", percentage)
	//		_, err := platform.Rest().UpdateMessage(call.ExtraProperties.ChannelId, message.ID, discord.MessageUpdate{
	//			Content: cohere.String(fmt.Sprintf("Downloading %v%%", percentage)),
	//		})
	//		if err != nil {
	//			panic(err)
	//		}
	//	},
	//	ProgressError: func(input string) {
	//		dlog.Log.Error("something wrong happened", "input", input)
	//	},
	//}
	//
	//dlog.Log.Info("call", "params", call.ToolCall.Parameters)
	//
	//information := call.ToolCall.Parameters["information"].(string)
	//if !strings.HasPrefix(information, "https://") {
	//	information = youtube.Search(information).Url
	//}
	//go y.Play(information)

	conn, problem := getConn(call)
	if problem {
		return
	}
	player := youtube.DefaultPlayer{
		Segments: &packets,
		Conn:     conn,
		Playing:  false,
	}
	youtube.AddPlayer(&player, call.ExtraProperties.GuildId)
	player.Run()

	go platform.HandleDeepgramVoicePackets(conn, call.ExtraProperties)

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

func getConn(call *cohere.CommandCall) (voice.Conn, bool) {
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
			return nil, true
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
			return nil, true
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
			return nil, true
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
			return nil, true
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
			return nil, true
		}
		dlog.Log.Info("wrote silent frame successfully")
	}
	return conn, false
}

func pause(call *cohere.CommandCall) {
	youtube.GetPlayer(call.ExtraProperties.GuildId).Pause()
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "paused successfully",
			},
		},
	}
}
func stop(call *cohere.CommandCall) {
	youtube.GetPlayer(call.ExtraProperties.GuildId).Stop()
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "stopped successfully",
			},
		},
	}
}
func resume(call *cohere.CommandCall) {
	youtube.GetPlayer(call.ExtraProperties.GuildId).Resume()
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "resumed successfully",
			},
		},
	}
}

func search(call *cohere.CommandCall) {
	data := youtube.Search(call.Parameters["information"].(string))
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success": true,
				//"Title":   data.FullTitle,
				//"Video Description": data.Description,
				//"Duration": data.DurationString,
				//"Channel":  data.Channel,
				"url": data.Url,
				//"Likes":    data.LikeCount,
				//"Views":    data.ViewCount,
			},
		},
	}
}
