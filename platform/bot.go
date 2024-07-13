package platform

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"os"
)

var client *bot.Client

var lavalinkClient = disgolink.New(snowflake.ID(1253286243685630024))

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
		//	dlog.Info("Triggered event", reflect.TypeOf(event))
		//}),

		bot.WithEventListenerFunc(dbReadyHandler),
		bot.WithEventListenerFunc(dbThreadCreateHandler),
		bot.WithEventListenerFunc(dbMessageCreateHandler),
		bot.WithEventListenerFunc(dbMessageUpdateHandler),

		bot.WithEventListenerFunc(botIsUpReadyHandler),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			dlog.Info("starting to query tracks")
			query := "ytsearch:Rick Astley - Never Gonna Give You Up"
			var toPlay *lavalink.Track

			lavalinkClient.BestNode().LoadTracksHandler(context.TODO(), query, disgolink.NewResultHandler(
				func(track lavalink.Track) {
					// Loaded a single track
					toPlay = &track

					dlog.Info("loaded a single track", "track", track)
				},
				func(playlist lavalink.Playlist) {
					// Loaded a playlist
				},
				func(tracks []lavalink.Track) {
					// Loaded a search result
					//dlog.Info("loaded a search result", "tracks", tracks)
					toPlay = &tracks[0]

				},
				func() {
					// nothing matching the query found
				},
				func(err error) {
					// something went wrong while loading the track
				},
			))
			// DisGo
			channelId := snowflake.MustParse("847908927554322436")
			guildId := snowflake.MustParse("847908927554322432")
			err = Client().UpdateVoiceState(context.TODO(), guildId, &channelId, false, false)
			if err != nil {
				dlog.Error("failed to update voice state", err)
				panic(err)
			}

			err := lavalinkClient.Player(guildId).Update(context.TODO(), lavalink.WithTrack(*toPlay))
			if err != nil {
				dlog.Error("failed to update voice player", err)
				panic(err)
			}
			dlog.Info("playing track", "track", toPlay)
		}),
		bot.WithEventListenerFunc(onVoiceStateUpdate),
		bot.WithEventListenerFunc(onVoiceServerUpdate),

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
		node, err := lavalinkClient.AddNode(context.TODO(), disgolink.NodeConfig{
			Name:      "My Discord bot", // a unique node name
			Address:   "buses.sleepyinsomniac.eu.org",
			Password:  "youshallnotpass",
			Secure:    false, // ws or wss
			SessionID: "",    // only needed if you want to resume a previous lavalink session
		})
		if err != nil {
			dlog.Error(err.Error())
			panic(err)
		}
		dlog.Info("connected to lavalinkNode", "node status", node.Status())
	}()

}

type Handler struct {
}

func (h Handler) OnEvent(event bot.Event) {
	dlog.Debug("update client")
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
	dlog.Info("disgo close successfully")
}

func onVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	// filter all non bot voice state updates out
	if event.VoiceState.UserID != Client().ApplicationID() {
		return
	}
	lavalinkClient.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
}

func onVoiceServerUpdate(event *events.VoiceServerUpdate) {
	lavalinkClient.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
}
