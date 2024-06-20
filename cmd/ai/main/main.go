package main

import (
	"fmt"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/fuad-daoud/discord-ai/discord"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/respeecher"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	Token string
)

func init() {
	Token = os.Getenv("TOKEN")
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	//slog.SetLogLoggerLevel(slog.LevelDebug)
}

func logRequest(r *http.Request) {
	uri := r.RequestURI
	method := r.Method
	fmt.Println("Got request!", method, uri)
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		fmt.Fprintf(w, "Hello! you've requested %s\n", r.URL.Path)
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		codeParams, ok := r.URL.Query()["code"]
		if ok && len(codeParams) > 0 {
			statusCode, _ := strconv.Atoi(codeParams[0])
			if statusCode >= 200 && statusCode < 600 {
				w.WriteHeader(statusCode)
			}
		}
	})

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

	//stop := make(chan os.Signal, 1)
	//<-stop
	//defer client.Close(context.TODO())
	//guildId := snowflake.MustParse(`847908927554322432`)
	//conn := client.VoiceManager().GetConn(guildId)
	//defer conn.Close(context.TODO())
	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		panic(err)
	}

	//signal.Notify(stop, os.Interrupt)
	log.Println("Graceful shutdown")
}
