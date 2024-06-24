package main

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/fuad-daoud/discord-ai/discord"
	"github.com/fuad-daoud/discord-ai/http"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/respeecher"
	"golang.org/x/net/context"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

var (
	Token string
)

func init() {
	Token = os.Getenv("TOKEN")
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	//slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	go http.SetupHttp()

	deepgramClient := deepgram.MakeDefault()
	respeecherClient := respeecher.MakeClient()

	client, err := disgo.New(Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentsNonPrivileged,
				gateway.IntentMessageContent,
				gateway.IntentGuildMembers,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentDirectMessageReactions,
				gateway.IntentDirectMessageTyping,
				gateway.IntentGuildVoiceStates,
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagsAll),
		),
		bot.WithEventListenerFunc(discord.BotIsUp),
		bot.WithEventListenerFunc(discord.VoiceServerUpdateHandler(deepgramClient)),
		bot.WithEventListenerFunc(discord.MessageCreateHandler(deepgramClient, respeecherClient)),
	)

	if err != nil {
		panic(err)
	}
	if err = client.OpenGateway(context.TODO()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	defer client.Close(context.TODO())
	slog.Info("Graceful shutdown")
}
