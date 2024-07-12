package commands

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/elevenlabs"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
	"golang.org/x/net/context"
)

func AddCommandsChannelOnReadyHandler() {
	go func() {
		for call := range cohere.Call {

			guildId := getGuildId(call.ExtraProperties.MessageId)
			call.ExtraProperties.GuildId = guildId

			switch call.Name {
			case "command_join":
				go joinFunction(call)
				break
			case "command_leave":
				go leave(call)
				break
			default:
				{
					cohere.Result <- &cohere.CommandResult{
						Call: call.ToolCall,
						Outputs: []map[string]interface{}{
							{
								"Success":     false,
								"Description": "command not implemented",
							},
						},
					}
					break
				}
			}
		}
	}()
}

func getGuildId(messageId string) string {
	m := cypher.MatchN("m", db.Message{Id: messageId})
	r := cypher.Return("g")
	result := db.Query(
		m,
		"-[:CONTAINS]-(t:Thread)-[:CHILD]-(c:TextChannel)-[:HAS]-(g:Guild)",
		r,
	)
	guild, parsed := cypher.ParseKey[db.Guild]("g", result)
	if !parsed {
		result = db.Query(
			m,
			"-[:CONTAINS]-(c:TextChannel)-[:HAS]-(g)",
			r,
		)
		guild, parsed = cypher.ParseKey[db.Guild]("g", result)
		if !parsed {
			panic("Can't find guild")
		}
	}
	return guild.Id
}

func joinFunction(call *cohere.CommandCall) {
	dlog.Info("starting join function")
	toolCall := call.ToolCall

	guildId := snowflake.MustParse(call.ExtraProperties.GuildId)
	userId := snowflake.MustParse(call.ExtraProperties.UserId)

	voiceState, b := platform.Cache().VoiceState(guildId, userId)
	if !b {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "user is not in a voice channel, the user shall be in a voice channel first",
				},
			},
		}
		return
	}

	botState, botStateOk := platform.Cache().VoiceState(guildId, platform.Client().ApplicationID())
	userState, userStateOk := platform.Cache().VoiceState(guildId, userId)

	if botStateOk && userStateOk && (botState.ChannelID.String() == userState.ChannelID.String()) {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "already in the voice channel",
				},
			},
		}
		return
	}

	conn := platform.Client().VoiceManager().CreateConn(guildId)
	dlog.Info("Staring joinVoiceChannel function")

	if err := conn.Open(context.Background(), *voiceState.ChannelID, false, false); err != nil {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error connecting to voice channel",
				},
			},
		}
		return
	}
	dlog.Info("opened connection successfully")
	if err := conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone); err != nil {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error setting speaking flag",
				},
			},
		}
		return
	}
	dlog.Info("set speaking successfully")
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		cohere.Result <- &cohere.CommandResult{
			Call: toolCall,
			Outputs: []map[string]interface{}{
				{
					"Success":     false,
					"Description": "error sending silence",
				},
			},
		}
		return
	}
	dlog.Info("wrote silent frame successfully")

	go platform.HandleDeepgramVoicePackets(conn, call.ExtraProperties.MessageId)

	cohere.Result <- &cohere.CommandResult{
		Call: toolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     false,
				"Description": "connected",
			},
		},
	}
	audioProvider, err := elevenlabs.TTS("I am here !!")
	if err != nil {
		//panic(err)
		return
	}

	conn.SetOpusFrameProvider(audioProvider)

	dlog.Info("Finished joining function")
	return
}
