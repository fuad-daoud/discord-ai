package platform

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/elevenlabs"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"net"
	"os"
	"strings"
	"time"
)

func HandleDeepgramVoicePackets(conn voice.Conn, messageId string) {

	dlog.Info("Added packets handler")
	guildID := conn.GuildID()
	result := db.Query(cypher.MatchN("m", db.Message{Id: messageId}), "-[]-", "(t:Thread)", cypher.Return("t"))
	thread, _ := cypher.ParseKey[db.TextChannel]("t", result)
	Cache().VoiceStatesForEach(guildID, func(state discord.VoiceState) {
		Cache().VoiceStatesForEach(guildID, func(state discord.VoiceState) {
			if state.ChannelID == conn.ChannelID() {
				deepgram.MakeClient(state.UserID.String(), finishedCallBack(conn, guildID, thread))
			}
		})
	})

	for {
		packet, err := conn.UDP().ReadPacket()
		if err != nil {
			return
		}
		voiceState, _ := Cache().VoiceState(guildID, Client().ApplicationID())
		//if !b {
		//	dlog.Error("bot not connected to a voice channel")
		//}
		if voiceState.SelfDeaf || voiceState.GuildDeaf {
			continue
		}
		userId := conn.UserIDBySSRC(packet.SSRC)
		deepgram.MakeClient(userId.String(), finishedCallBack(conn, guildID, thread))
		deepgram.Write(packet.Opus, userId.String())
	}
}

func finishedCallBack(conn voice.Conn, guildId snowflake.ID, thread db.TextChannel) deepgram.FinishedCallBack {

	return func(message string, userId string) {
		dlog.Info("finished call back starting ...", "userId", userId)
		snowflakeUserId := snowflake.MustParse(userId)
		user, err := Rest().GetUser(snowflakeUserId)
		if err != nil {
			dlog.Error("could not get user: ", userId, err)
		}
		if !(isCallingMe(message)) {
			return
		}

		userState, userStateOk := Cache().VoiceState(guildId, snowflakeUserId)
		if !userStateOk {
			dlog.Error("Member voice state not okay")
		}
		err = Client().UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, true)
		if err != nil {
			dlog.Error("could not update voice state: ", err)
		}

		//MATCH (m:Message {id: "1255887883848646769"}), (t:Thread) MATCH shortestPath((m)-[*]->(t)) RETURN m,t
		messageId := handleThread(thread.Id, *user, message)
		response := cohere.Send(message, messageId, userId, thread.Id)

		audioProvider, err := elevenlabs.TTS(response)

		selfUser, b := Cache().SelfUser()
		if !b {
			dlog.Error("could not get self user")
		}
		go handleThread(thread.Id, selfUser.User, response)
		conn.SetOpusFrameProvider(audioProvider)
		err = Client().UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, false)
		if err != nil {
			dlog.Error("could not update voice state: ", err)
		}
	}
}

func handleThread(threadId string, user discord.User, message string) string {
	message = fmt.Sprintf("%s: %s", user.Username, message)
	createMessage, err := Rest().CreateMessage(snowflake.MustParse(threadId), discord.MessageCreate{Content: message})
	if err != nil {
		panic(err)
	}
	dlog.Info("Created message", "ID", createMessage.ID)
	return createMessage.ID.String()
}

func messageCreateHandler(event *events.GuildMessageCreate) {
	authorId := event.Message.Author.ID
	if authorId == event.Client().ID() {
		return
	}
	restClient := event.Client().Rest()
	channel, err := restClient.GetChannel(event.ChannelID)
	if err != nil {
		return
	}
	dlog.Info("got channel", "channel", channel.Name())

	messageContent := event.Message.Content
	if channel.Type() == discord.ChannelTypeGuildPublicThread {
		var process Process
		thread := channel.(discord.GuildThread)
		dlog.Info("Got thread", "ID", thread.ID())
		if thread.ParentID().String() == "1252536839886082109" {
			process = func(message, messageId, memberId, threadId string) string {
				botState, botStateOk := event.Client().Caches().VoiceState(event.GuildID, event.Client().ApplicationID())

				_, userStateOk := event.Client().Caches().VoiceState(event.GuildID, authorId)
				if !userStateOk {
					return "You are not in a voice channel bro "
				}
				go func() {
					err := event.Client().Rest().SendTyping(thread.ID())
					if err != nil {
						panic(err)
					}
				}()

				if botStateOk {
					err := deafen(&event.GuildID, botState.ChannelID)
					if err != nil {
						panic(err)
					}
				}
				response := cohere.Send(message+"(respond like you are whispering)", event.MessageID.String(), authorId.String(), thread.ID().String())
				audioProvider, err := elevenlabs.TTS(response)
				if err != nil {
					dlog.Error("Failed to send speech", "err", err)
					panic(err)
				}
				if botStateOk {
					conn := event.Client().VoiceManager().GetConn(event.GuildID)
					conn.SetOpusFrameProvider(audioProvider)
					return response
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				conn := event.Client().VoiceManager().CreateConn(event.GuildID)
				err = conn.Open(ctx, *botState.ChannelID, false, false)
				if err != nil {
					dlog.Error("Failed to open voice channel", "channel", event.GuildID, "error", err)
					panic(err)
				}
				conn.SetOpusFrameProvider(audioProvider)
				return response
			}
		} else {
			process = cohere.Send
		}
		replyText(thread.ID(), messageContent, event.MessageID.String(), authorId.String(), process)
	} else {
		if channel.ID().String() != "1252273230727876619" && channel.ID().String() != "1252536839886082109" && channel.ID().String() != "1256856679379636276" {
			return
		}

		if !(isCallingMe(messageContent)) {
			return
		}
		var threadName string

		if strings.Index(messageContent, "\n") == -1 {
			if len(messageContent) > 10 {
				threadName = messageContent[0:10]
			} else {
				threadName = messageContent
			}
		} else {
			threadName = messageContent[0:strings.Index(messageContent, "\n")]
		}

		newThread, err := restClient.CreateThreadFromMessage(channel.ID(), event.MessageID, discord.ThreadCreateFromMessage{Name: threadName, AutoArchiveDuration: 1440})
		if err != nil {
			dlog.Error("could not create discord thread", err.Error())
			panic(err)
		}

		replyText(newThread.ID(), messageContent, event.MessageID.String(), authorId.String(), cohere.Send)
	}
}

func botIsUpReadyHandler(event *events.Ready) {
	user, _ := event.Client().Caches().SelfUser()
	dlog.Info("Bot is up!")
	dlog.Info("Bot", "username", user.Username)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ips, err := GetLocalIPs()
	if err != nil {
		panic(err)
	}
	content := "I AM ALIVE in " + hostname + " IPs [ "
	for _, ip := range ips {
		content += ip.String() + " "
	}
	content += "]"
	message, err := event.Client().Rest().CreateMessage(snowflake.MustParse("1252273230727876619"), discord.MessageCreate{
		Content: content,
	})
	if err != nil {
		panic(err)
	}
	dlog.Info("Created message", "ID", message.ID.String(), "content", message.Content)
}
func GetLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

func replyText(channelId snowflake.ID, content, messageId, authorId string, process Process) {
	processingMessage := "Dazzling✨💫"

	message, err := Rest().CreateMessage(channelId, discord.MessageCreate{
		Content: processingMessage,
	})
	if err != nil {
		panic(err)
	}

	response := process(content, messageId, authorId, channelId.String())
	updateMessage, err := Rest().UpdateMessage(channelId, message.ID, discord.MessageUpdate{Content: &response})
	if err != nil {
		panic(err)
	}
	dlog.Info("updated message:", "ID", updateMessage.ID.String())
}

type Process func(message, messageId, userId, threadId string) string

func voiceServerUpdateHandler(event *events.GuildVoiceStateUpdate) {
	if event.Member.User.ID == Client().ID() {
		dlog.Info("Update on bot voice state")
		id := event.GenericGuildVoiceState.VoiceState.ChannelID
		if id == nil {
			dlog.Info("Disconnected from voice channel")
			deepgram.Stop()
			return
		}
		return
	}
}

func isCallingMe(message string) bool {
	message = strings.ToLower(message)
	prefixes := []string{"luna", "hey luna", "hello luna", "hello, luna", "you luna", "ya luna", "ola luna", "luna hello", "luna, hello"}
	dlog.Info("detecting message", "message", message)
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}
