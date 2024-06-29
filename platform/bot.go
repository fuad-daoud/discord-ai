package platform

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"golang.org/x/net/context"
	"log/slog"
	"os"
	"time"
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

		//bot.WithEventListenerFunc(func(event *events.GenericEvent) {
		//	slog.Info("Triggered event", reflect.TypeOf(event))
		//}),

		bot.WithEventListenerFunc(dbReadyHandler),
		bot.WithEventListenerFunc(dbThreadCreateHandler),
		bot.WithEventListenerFunc(dbMessageCreateHandler),
		bot.WithEventListenerFunc(dbMessageUpdateHandler),

		bot.WithEventListenerFunc(botIsUpReadyHandler),
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

	go func() {
		time.Sleep(2 * time.Second)
		addCommandsChannelOnReadyHandler()

		//gpt.Action <- gpt.FunctionInput{
		//	Function: gpt.Function{
		//		Name:      "join",
		//		Arguments: "",
		//	},
		//	UserId:    "468494540852953089",
		//	MessageId: "1255909408043696190",
		//}

	}()

}

type Handler struct {
}

func (h Handler) OnEvent(event bot.Event) {
	slog.Info("update client")
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
	slog.Info("disgo close successfully")
}
