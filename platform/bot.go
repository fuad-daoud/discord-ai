package platform

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/fuad-daoud/discord-ai/integrations/lava"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"os"
)

var client *bot.Client

func Setup() {
	var err error
	clientTmp, err := disgo.New(os.Getenv("TOKEN"),
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

		bot.WithEventListenerFunc(dbReadyHandler),
		bot.WithEventListenerFunc(dbThreadCreateHandler),
		bot.WithEventListenerFunc(dbMessageCreateHandler),
		bot.WithEventListenerFunc(dbMessageUpdateHandler),

		bot.WithEventListenerFunc(botIsUpReadyHandler),
		//bot.WithEventListenerFunc(lava.OnVoiceServerUpdate),
		//bot.WithEventListenerFunc(lava.OnVoiceStateUpdate),
		bot.WithEventListenerFunc(lava.OnReady),

		bot.WithEventListenerFunc(func(e *events.GuildMessageCreate) {
			go messageCreateHandler(e)
		}),
		bot.WithEventListenerFunc(voiceServerUpdateHandler),

		//bot.WithEventListenerFunc(addCommandsChannelOnReadyHandler),
	)
	if err != nil {
		panic(err)
	}
	if err = clientTmp.OpenGateway(context.TODO()); err != nil {
		panic(err)
	}
	clientTmp.EventManager().AddEventListeners(Handler{})
	client = &clientTmp
}

type Handler struct {
}

func (h Handler) OnEvent(event bot.Event) {
	dlog.Log.Debug("update client")
	c := event.Client()
	client = &c
}

func Client() bot.Client {
	if client != nil {
		return *client
	}
	Setup()
	return *client
}

func Rest() rest.Rest {
	return Client().Rest()
}
func Cache() cache.Caches {
	return Client().Caches()
}

func Close() {
	(*client).Close(context.TODO())
	dlog.Log.Info("disgo close successfully")
}
