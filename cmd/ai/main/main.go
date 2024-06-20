package main

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/discord"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/respeecher"
	"golang.org/x/net/context"
	"log"
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
			//cache.WithMemberCache(cache.NewMemberCache(newGroupedCache[discord.Member]())),
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
	<-stop
	defer client.Close(context.TODO())
	guildId := snowflake.MustParse(`847908927554322432`)
	conn := client.VoiceManager().GetConn(guildId)
	defer conn.Close(context.TODO())
	signal.Notify(stop, os.Interrupt)
	log.Println("Graceful shutdown")
}
