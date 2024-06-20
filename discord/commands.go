package discord

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/gpt"
	"github.com/fuad-daoud/discord-ai/integrations/respeecher"
	"golang.org/x/net/context"
	"log/slog"
	"time"
)

func CommandsProcessor(chanRun chan gpt.Run, client bot.Client, gptClient gpt.Client, deepgramClient deepgram.Client, respeecherClient respeecher.Client) {
	for run := range chanRun {
		if run.Status != "requires_action" {
			continue
		}
		functionName := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Function.Name
		switch functionName {
		case "join":
			JoinFunction(chanRun, client, gptClient, deepgramClient, respeecherClient, run)
			break
		}
	}
}

func JoinFunction(chanRun chan gpt.Run, client bot.Client, gptClient gpt.Client, deepgramClient deepgram.Client, respeecherClient respeecher.Client, run gpt.Run) {
	slog.Info("starting join function")
	toolCallId := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Id
	guildId := snowflake.MustParse(`847908927554322432`)
	userId := snowflake.MustParse(run.MetaData.UserId)
	voiceState, b := client.Caches().VoiceState(guildId, userId)
	if !b {
		slog.Error("could not get voice state")
	}

	botState, botStateOk := client.Caches().VoiceState(guildId, client.ApplicationID())
	userState, userStateOk := client.Caches().VoiceState(guildId, userId)

	if botStateOk && userStateOk && botState.ChannelID.String() == userState.ChannelID.String() {
		newRun := gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
			Success:     "false",
			Description: "already in the voice channel",
		})
		chanRun <- *newRun
		return
	}
	conn := client.VoiceManager().CreateConn(guildId)
	go joinAndPlay(conn, run, deepgramClient, respeecherClient, gptClient, client, voiceState.ChannelID)

	newRun := gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
		Success:     "true",
		Description: "",
	})

	chanRun <- *newRun
	return
}

func joinAndPlay(conn voice.Conn, run gpt.Run, deepgramClient deepgram.Client, respeecherClient respeecher.Client, gptClient gpt.Client, client bot.Client, channelId *snowflake.ID) {
	slog.Info("Staring joinAndPlay function")
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
	defer cancel()
	if err := conn.Open(ctx, *channelId, false, false); err != nil {
		panic("error connecting to voice channel: " + err.Error())
	}
	slog.Info("opened connection successfully")
	if err := conn.SetSpeaking(ctx, voice.SpeakingFlagMicrophone); err != nil {
		panic("error setting speaking flag: " + err.Error())
	}
	slog.Info("set speaking successfully")
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		panic("error sending silence: " + err.Error())
	}
	slog.Info("wrote silent frame successfully")

	guildId := snowflake.MustParse(`847908927554322432`)

	slog.Info("starting playback")
	err := Talk(conn, "files/fixed-replies/hey-luna-is-here-oksana.wav", func() error {
		return client.UpdateVoiceState(context.Background(), guildId, channelId, false, true)
	}, func() error {
		return client.UpdateVoiceState(context.Background(), guildId, channelId, false, false)
	})
	if err != nil {
		slog.Info("error talking to voice channel:", "err", err)
	}
	data := run.MetaData
	go handleDeepgramVoicePackets(conn, deepgramClient, finishedCallBack(conn, client, gptClient, respeecherClient, data.ChannelId), client)
}
