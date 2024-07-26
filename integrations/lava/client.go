package lava

import (
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
)

var client disgolink.Client

func Client() disgolink.Client {
	return client
}

func OnReady(event *events.Ready) {
	go func() {
		//client = disgolink.New(event.Application.ID)
		//node, err := client.AddNode(context.TODO(), disgolink.NodeConfig{
		//	Name:      "My Discord bot", // a unique node name
		//	Address:   "buses.sleepyinsomniac.eu.org",
		//	Password:  "youshallnotpass",
		//	Secure:    false, // ws or wss
		//	SessionID: "",    // only needed if you want to resume a previous lava session
		//})
		//if err != nil {
		//	dlog.Log.Error(err.Error())
		//	panic(err)
		//}
		//
		//dlog.Log.Info("connected to lavalinkNode", "node status", node.Status())
	}()
}

func OnVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.UserID != event.Client().ApplicationID() {
		return
	}
	client.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
}

func OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	//conn := event.Client().VoiceManager().GetConn(event.GuildID)
	//if conn == nil {
	player := client.Player(event.GuildID)

	//_, err := player.Node().Rest().UpdatePlayer(context.TODO(), player.Node().SessionID(), player.GuildID(), lavalink.PlayerUpdate{
	//	Voice: &lavalink.VoiceState{
	//		Token:     event.Token,
	//		Endpoint:  *event.Endpoint,
	//		SessionID: "my_session_id",
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}

	dlog.Log.Info("player", "player session id", player.Node().SessionID())
	client.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
	//}
}
