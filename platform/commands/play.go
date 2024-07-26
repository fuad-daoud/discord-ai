package commands

import (
	"context"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/digitalocean"
	"github.com/fuad-daoud/discord-ai/integrations/youtube"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
)

func play(call *cohere.CommandCall) {
	dlog.Log.Info("starting play function")
	dlog.Log.Info("play call", "params", call.ToolCall.Parameters)
	toolCall := call.ToolCall
	information := call.ToolCall.Parameters["information"].(string)
	data := youtube.Search(information)
	packets := getPackets(call, data)

	conn, problem := getConn(call)
	if problem {
		return
	}
	var player youtube.Player

	player = youtube.GetPlayer(call.ExtraProperties.GuildId)

	player.SetConn(conn)
	player.Add(data, packets)
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

func getPackets(call *cohere.CommandCall, data youtube.Data) *[][]byte {
	var packets *[][]byte
	dlog.Log.Info("Checking cache")
	download := digitalocean.Download("youtube/cache/" + data.Id + ".opus")
	if download != nil {
		dlog.Log.Info("cache found")
		packets = audio.ReadDCA(download.Body)
	} else {
		dlog.Log.Info("no cache ..")
		message, err := platform.Rest().CreateMessage(call.ExtraProperties.ChannelId, discord.MessageCreate{
			Content: "Downloading ...",
		})
		if err != nil {
			panic(err)
		}
		y := youtube.Ytdlp{
			Progress:      progress(call, message),
			ProgressError: progressError(),
			Data:          data,
		}

		packets, err = y.GetAudio()
		if err != nil {
			panic(err)
		}
	}
	return packets
}

func progress(call *cohere.CommandCall, message *discord.Message) func(percentage float64) {
	return func(percentage float64) {
		dlog.Log.Info("Progress ", "percentage", percentage)
		_, err := platform.Rest().UpdateMessage(call.ExtraProperties.ChannelId, message.ID, discord.MessageUpdate{
			Content: cohere.String(fmt.Sprintf("Downloading %v%%", percentage)),
		})
		if err != nil {
			panic(err)
		}
	}
}

func progressError() func(input string) {
	return func(input string) {
		dlog.Log.Error("something wrong happened", "input", input)
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
func skip(call *cohere.CommandCall) {
	youtube.GetPlayer(call.ExtraProperties.GuildId).Skip()
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     true,
				"Description": "skipped successfully",
			},
		},
	}
}
func queue(call *cohere.CommandCall) {
	player := youtube.GetPlayer(call.ExtraProperties.GuildId)
	q := youtube.GetQueue(player.GetDBPlayer())
	result := make([]map[string]interface{}, 0)
	for _, element := range q {
		result = append(result, map[string]interface{}{
			"FullTitle":      element.FullTitle,
			"DurationString": element.DurationString,
			"Url":            element.Url,
		})
	}
	cohere.Result <- &cohere.CommandResult{
		Call: call.ToolCall,
		Outputs: []map[string]interface{}{
			{
				"Success": true,
				"Queue":   result,
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
