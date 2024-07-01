package platform

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
	"github.com/fuad-daoud/discord-ai/integrations/gpt"
	"golang.org/x/net/context"
	"log/slog"
)

func JoinFunction(input gpt.FunctionInput) {
	slog.Info("starting join function")
	guild := getGuild(input)
	guildId := snowflake.MustParse(guild.Id)
	userId := snowflake.MustParse(input.UserId)

	voiceState, b := Cache().VoiceState(guildId, userId)
	if !b {
		gpt.Response <- gpt.FunctionOutput{
			Success:     false,
			Description: "you are not in a voice channel stupid",
		}
		return
	}

	botState, botStateOk := Cache().VoiceState(guildId, Client().ApplicationID())
	userState, userStateOk := Cache().VoiceState(guildId, userId)

	if botStateOk && userStateOk && botState.ChannelID.String() == userState.ChannelID.String() {
		gpt.Response <- gpt.FunctionOutput{
			Success:     false,
			Description: "already in the voice channel",
		}
		return
	}

	conn := Client().VoiceManager().CreateConn(guildId)
	slog.Info("Staring joinVoiceChannel function")

	if err := conn.Open(context.Background(), *voiceState.ChannelID, false, false); err != nil {
		gpt.Response <- gpt.FunctionOutput{
			Success:     false,
			Description: "error connecting to voice channel",
		}
		return
	}
	slog.Info("opened connection successfully")
	if err := conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone); err != nil {
		gpt.Response <- gpt.FunctionOutput{
			Success:     false,
			Description: "error setting speaking flag",
		}
		return
	}
	slog.Info("set speaking successfully")
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		gpt.Response <- gpt.FunctionOutput{
			Success:     false,
			Description: "error sending silence",
		}
		return
	}
	slog.Info("wrote silent frame successfully")

	slog.Info("starting playback")

	go handleDeepgramVoicePackets(conn, input.MessageId)

	gpt.Response <- gpt.FunctionOutput{
		Success:     true,
		Description: "connected",
	}
	slog.Info("Finished joining function")
	return
}

func getGuild(input gpt.FunctionInput) db.Guild {
	m := cypher.MatchN("m", db.Message{Id: input.MessageId})
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
	return guild
}
