package commands

import (
	"bufio"
	"bytes"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/platform"
	"golang.org/x/net/context"
	"io/ioutil"
	"log/slog"
)

func AddCommandsChannelOnReadyHandler() {
	go func() {
		for call := range cohere.Call {
			messageId, ok := call.Properties["messageId"].(string)
			if ok {
				guildId := getGuildId(messageId)
				call.Properties["guildId"] = guildId
			}

			switch call.Name {
			case "command_join":
				go joinFunction(call)
				break
			case "command_leave":
				go leave(call)
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
	slog.Info("starting join function")
	toolCall := call.ToolCall
	properties := call.Properties
	guildId := snowflake.MustParse(properties["guildId"].(string))
	userId := snowflake.MustParse(properties["userId"].(string))

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
	slog.Info("Staring joinVoiceChannel function")

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
	slog.Info("opened connection successfully")
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
	slog.Info("set speaking successfully")
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
	slog.Info("wrote silent frame successfully")

	go platform.HandleDeepgramVoicePackets(conn, properties["messageId"].(string))

	cohere.Result <- &cohere.CommandResult{
		Call: toolCall,
		Outputs: []map[string]interface{}{
			{
				"Success":     false,
				"Description": "connected",
			},
		},
	}
	slog.Info("Testing talking")
	file, _ := ioutil.ReadFile("/home/fuad/GolandProjects/discord-ai/output.opus")
	closer := bytes.NewReader(file)
	conn.SetOpusFrameProvider(&platform.AudioProvider{
		Source: bufio.NewReader(closer),
	})
	slog.Info("Finished joining function")
	return
}
