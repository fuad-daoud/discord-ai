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
	"golang.org/x/net/context"
	"log/slog"
	"strings"
	"time"
)

func HandleDeepgramVoicePackets(conn voice.Conn, messageId string) {

	slog.Info("Added packets handler")
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
		//	slog.Error("bot not connected to a voice channel")
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
		slog.Info("finished call back starting ...", "userId", userId)
		snowflakeUserId := snowflake.MustParse(userId)
		user, err := Rest().GetUser(snowflakeUserId)
		if err != nil {
			slog.Error("could not get user: ", userId, err)
		}
		if !(strings.HasPrefix(message, "luna") || strings.HasPrefix(message, "Luna")) {
			return
		}

		userState, userStateOk := Cache().VoiceState(guildId, snowflakeUserId)
		if !userStateOk {
			slog.Error("Member voice state not okay")
		}
		err = Client().UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, true)
		if err != nil {
			slog.Error("could not update voice state: ", err)
		}

		//MATCH (m:Message {id: "1255887883848646769"}), (t:Thread) MATCH shortestPath((m)-[*]->(t)) RETURN m,t
		go handleThread(thread.Id, *user, message)
		response := cohere.Send(message, "", userId, thread.Id)

		voiceReader, err := deepgram.TTS(response)
		if err != nil {
			slog.Error("could not send tts: ", err)
			panic(err)
		}

		selfUser, b := Cache().SelfUser()
		if !b {
			slog.Error("could not get self user")
		}
		go handleThread(thread.Id, selfUser.User, response)
		conn.SetOpusFrameProvider(&AudioProvider{
			Source: voiceReader,
		})
		err = Client().UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, false)
		if err != nil {
			slog.Error("could not update voice state: ", err)
		}
	}
}

//go indicator(voiceConnection, "processing-this-wont-take-long.wav")
//go indicator(conn, "rizz-sounds.mp3", )
//go indicator(voiceConnection, "formula-1-radio-notification.mp3")

// text := "Hey there, I'm Luna, your stunning Discord bot. What can I do for you today?"
// text := "I don't have access to real-time information like the current date. But you can check on your device, love. Is there anything else you'd like me to do?"
//
//	func indicator(conn voice.Conn, file string) {
//		err := Talk(conn, "files/fixed-replies/"+file)
//		if err != nil {
//			panic(err)
//		}
//	}
func handleThread(threadId string, user discord.User, message string) {
	message = fmt.Sprintf("%s: %s", user.Mention(), message)
	createMessage, err := Rest().CreateMessage(snowflake.MustParse(threadId), discord.MessageCreate{Content: message})
	if err != nil {
		panic(err)
	}
	slog.Info("Created message", "ID", createMessage.ID)
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
	slog.Info("got channel", "channel", channel.Name())

	messageContent := event.Message.Content
	if channel.Type() == discord.ChannelTypeGuildPublicThread {
		var process Process
		thread := channel.(discord.GuildThread)
		slog.Info("Got thread", "ID", thread.ID())
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
				voiceReader, err := deepgram.TTS(response)
				if err != nil {
					slog.Error("Failed to send speech", "err", err)
					panic(err)
				}
				if botStateOk {
					conn := event.Client().VoiceManager().GetConn(event.GuildID)
					conn.SetOpusFrameProvider(&AudioProvider{
						Source: voiceReader,
					})
					return response
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				conn := event.Client().VoiceManager().CreateConn(event.GuildID)
				err = conn.Open(ctx, *botState.ChannelID, false, false)
				if err != nil {
					slog.Error("Failed to open voice channel", "channel", event.GuildID, "error", err)
					panic(err)
				}
				conn.SetOpusFrameProvider(&AudioProvider{
					Source: voiceReader,
				})
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

		if !(strings.HasPrefix(messageContent, "luna") || strings.HasPrefix(messageContent, "Luna")) {
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
			slog.Error("could not create discord thread", err.Error())
			panic(err)
		}
		slog.Info("Created discord thread with gpt id as name", "name", newThread.Name(), "id", newThread.ID())

		//response := cohere.Send(event.Message.Content, event.MessageID.String(), authorId.String(), newThread.ID().String())
		//response := gpt.SendMessageFullCycle(event.Message.Content, event.MessageID.String(), authorId.String(), gptThread.Id)

		replyText(newThread.ID(), messageContent, event.MessageID.String(), authorId.String(), cohere.Send)
	}
}

func botIsUpReadyHandler(event *events.Ready) {
	user, _ := event.Client().Caches().SelfUser()
	slog.Info("Bot is up!")
	slog.Info("Bot", "username", user.Username)
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
	slog.Info("updated message:", "ID", updateMessage.ID.String())
}

type Process func(message, messageId, userId, threadId string) string

func voiceServerUpdateHandler(event *events.GuildVoiceStateUpdate) {
	if event.Member.User.ID == Client().ID() {
		slog.Info("Update on bot voice state")
		id := event.GenericGuildVoiceState.VoiceState.ChannelID
		if id == nil {
			slog.Info("Disconnected from voice channel")
			deepgram.Stop()
			return
		}
		return
	}
}
