package platform

import (
	"bufio"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/integrations/cohere"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/youtube"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"net"
	"os"
	"strings"
)

func finishedCallBack(conn voice.Conn, properties cohere.Properties) deepgram.FinishedCallBack {

	return func(message string, userId string) {
		dlog.Log.Info("finished call back starting ...", "userId", userId)
		snowflakeUserId := snowflake.MustParse(userId)
		user, err := Rest().GetUser(snowflakeUserId)
		if err != nil {
			dlog.Log.Error("could not get user: ", userId, err)
		}
		if !(isCallingMe(message)) {
			return
		}

		userState, userStateOk := Cache().VoiceState(properties.GuildId, snowflakeUserId)
		if !userStateOk {
			dlog.Log.Error("Member voice state not okay")
		}
		err = Client().UpdateVoiceState(context.Background(), properties.GuildId, userState.ChannelID, false, true)
		if err != nil {
			dlog.Log.Error("could not update voice state: ", err)
		}

		//MATCH (m:Message {id: "1255887883848646769"}), (t:Thread) MATCH shortestPath((m)-[*]->(t)) RETURN m,t
		handleThread(properties.ChannelId, *user, message)

		processingMessage := "Dazzling✨💫"

		selfUser, b := Cache().SelfUser()
		if !b {
			dlog.Log.Error("could not get self user")
		}
		messageId := handleThread(properties.ChannelId, selfUser.User, processingMessage)

		//err = Rest().AddReaction(channelId, message.ID, "🌙")
		//if err != nil {
		//	panic(err)
		//}

		streamResult := cohere.StreamChat(message, properties.ChannelId.String(), cohere.Properties{
			MessageId: properties.MessageId,
			UserId:    snowflake.MustParse(userId),
			GuildId:   properties.GuildId,
			ChannelId: properties.ChannelId,
		})

		byteLength := 0
		byteString := make([]byte, 1000)
		newBytes := 0
		for result := range streamResult {
			dlog.Log.Debug("got result", "result type", result.Type)
			switch result.Type {
			case cohere.Start:
				{

					updateMessage, err := Rest().UpdateMessage(properties.ChannelId, messageId, discord.MessageUpdate{Content: cohere.String("Hmmm🤔... ")})
					if err != nil {
						panic(err)
					}
					err = Rest().AddReaction(properties.ChannelId, messageId, "🔵")
					if err != nil {
						panic(err)
					}
					dlog.Log.Debug("started message:", "ID", updateMessage.ID.String())
					break
				}
			case cohere.Text:
				{
					go func() {
						copiedBytes := copy(byteString[byteLength:], result.Message)
						newBytes += copiedBytes
						byteLength += copiedBytes
						if newBytes < 20 {
							return
						}
						newBytes = 0

						_, err := Rest().UpdateMessage(properties.ChannelId, messageId, discord.MessageUpdate{Content: cohere.String(string(byteString))})
						if err != nil {
							panic(err)
						}
						//dlog.Log.Debug("updated message:", "ID", updateMessage.ID.String())
					}()
					break
				}
			case cohere.End:
				{
					updateMessage, err := Rest().UpdateMessage(properties.ChannelId, messageId, discord.MessageUpdate{Content: cohere.String(result.Message)})
					if err != nil {
						panic(err)
					}
					dlog.Log.Debug("updated message:", "ID", updateMessage.ID.String())

					err = Rest().RemoveOwnReaction(properties.ChannelId, messageId, "🔵")
					if err != nil {
						panic(err)
					}
					err = Rest().AddReaction(properties.ChannelId, messageId, "🟢")
					if err != nil {
						panic(err)
					}
					dlog.Log.Debug("finished message:", "ID", messageId)

					err = Client().UpdateVoiceState(context.Background(), properties.GuildId, userState.ChannelID, false, false)

					return
				}
			}
		}

		//audioProvider, err := elevenlabs.TTS(response)

		//conn.SetOpusFrameProvider(audioProvider)

		if err != nil {
			dlog.Log.Error("could not update voice state: ", err)
		}
	}
}

func HandleDeepgramVoicePackets(conn voice.Conn, props cohere.Properties) {

	dlog.Log.Debug("Added packets handler")
	guildID := conn.GuildID()

	for {
		packet, err := conn.UDP().ReadPacket()
		if err != nil {
			return
		}
		voiceState, _ := Cache().VoiceState(guildID, Client().ApplicationID())
		//if !b {
		//	dlog.Log.Error("bot not connected to a voice channel")
		//}
		if voiceState.SelfDeaf || voiceState.GuildDeaf {
			continue
		}
		userId := conn.UserIDBySSRC(packet.SSRC)
		if userId.String() == "0" {
			continue
		}
		deepgram.MakeClient(userId.String(), finishedCallBack(conn, props))
		deepgram.Write(packet.Opus, userId.String())
	}
}

func handleThread(threadId snowflake.ID, user discord.User, message string) snowflake.ID {
	message = fmt.Sprintf("%s: %s", user.Username, message)
	createMessage, err := Rest().CreateMessage(threadId, discord.MessageCreate{Content: message})
	if err != nil {
		panic(err)
	}
	dlog.Log.Info("Created message", "ID", createMessage.ID)
	return createMessage.ID
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
	dlog.Log.Info("got channel", "channel", channel.Name())

	messageContent := event.Message.Content
	if channel.Type() == discord.ChannelTypeGuildPublicThread {
		thread := channel.(discord.GuildThread)
		dlog.Log.Info("Got thread", "ID", thread.ID())
		streamMessage(thread.ID(), messageContent, cohere.Properties{
			MessageId: event.MessageID,
			UserId:    authorId,
			GuildId:   event.GuildID,
			ChannelId: thread.ID(),
		})
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
			dlog.Log.Error("could not create discord thread", err.Error())
			panic(err)
		}
		streamMessage(newThread.ID(), messageContent, cohere.Properties{
			MessageId: event.MessageID,
			UserId:    authorId,
			GuildId:   event.GuildID,
			ChannelId: newThread.ID(),
		})
	}
}
func banner() string {
	readFile, err := os.Open("./assets/banner.txt")
	defer readFile.Close()
	if err != nil {
		dlog.Log.Error(err.Error())
		panic(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	builder := strings.Builder{}
	for fileScanner.Scan() {
		builder.WriteString(fileScanner.Text())
	}

	return builder.String()
}
func botIsUpReadyHandler(event *events.Ready) {
	user, _ := event.Client().Caches().SelfUser()
	dlog.Log.Info("Bot", "username", user.Username)
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ips, err := GetLocalIPs()
	if err != nil {
		panic(err)
	}
	content := hostname + " IPs [ "
	for _, ip := range ips {
		content += ip.String() + " "
	}
	content += "]"
	message, err := event.Client().Rest().CreateMessage(snowflake.MustParse("1266536257568116736"), discord.MessageCreate{
		Content: content,
	})
	if err != nil {
		panic(err)
	}
	dlog.Log.Info("Created message", "ID", message.ID.String(), "content", message.Content)
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

func streamMessage(channelId snowflake.ID, content string, prop cohere.Properties) {
	processingMessage := "Dazzling✨💫"

	message, err := Rest().CreateMessage(channelId, discord.MessageCreate{
		Content: processingMessage,
	})
	if err != nil {
		panic(err)
	}

	streamResult := cohere.StreamChat(content, channelId.String(), prop)

	err = Rest().AddReaction(channelId, message.ID, "🌙")
	if err != nil {
		panic(err)
	}

	byteLength := 0
	byteString := make([]byte, 1000)
	newBytes := 0
	for result := range streamResult {
		dlog.Log.Debug("got result", "result type", result.Type)
		switch result.Type {
		case cohere.Start:
			{
				updateMessage, err := Rest().UpdateMessage(channelId, message.ID, discord.MessageUpdate{Content: cohere.String("Thinking🤔...")})
				if err != nil {
					panic(err)
				}
				err = Rest().AddReaction(channelId, message.ID, "🔵")
				if err != nil {
					panic(err)
				}
				dlog.Log.Debug("started message:", "ID", updateMessage.ID.String())
				break
			}
		case cohere.Text:
			{
				go func() {
					copiedBytes := copy(byteString[byteLength:], result.Message)
					newBytes += copiedBytes
					byteLength += copiedBytes
					if newBytes < 20 {
						return
					}
					newBytes = 0

					_, err := Rest().UpdateMessage(channelId, message.ID, discord.MessageUpdate{Content: cohere.String(string(byteString))})
					if err != nil {
						panic(err)
					}
					//dlog.Log.Debug("updated message:", "ID", updateMessage.ID.String())
				}()
				break
			}
		case cohere.End:
			{
				updateMessage, err := Rest().UpdateMessage(channelId, message.ID, discord.MessageUpdate{Content: cohere.String(result.Message)})
				if err != nil {
					panic(err)
				}
				dlog.Log.Debug("updated message:", "ID", updateMessage.ID.String())

				err = Rest().RemoveOwnReaction(channelId, message.ID, "🔵")
				if err != nil {
					panic(err)
				}
				err = Rest().AddReaction(channelId, message.ID, "🟢")
				if err != nil {
					panic(err)
				}
				dlog.Log.Debug("finished message:", "ID", message.ID)
				return
			}
		}
	}
}

type Process func(message, messageId, userId, threadId string) string

func voiceServerUpdateHandler(event *events.GuildVoiceStateUpdate) {
	if event.Member.User.ID == Client().ID() {
		dlog.Log.Debug("Update on bot voice state")
		id := event.GenericGuildVoiceState.VoiceState.ChannelID
		if id == nil {
			dlog.Log.Info("Disconnected from voice channel")
			deepgram.Stop()
			player, _ := youtube.GetPlayer(event.OldVoiceState.GuildID)
			if player != nil {
				player.Idle()
			}
			return
		}
		return
	}
}

func isCallingMe(message string) bool {
	message = strings.ToLower(message)
	prefixes := []string{"luna", "hey luna", "hello luna", "hello, luna", "you luna", "ya luna", "ola luna", "luna hello", "luna, hello", "luan", "Luan"}
	dlog.Log.Debug("detecting message", "message", message)
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	suffixes := []string{"luna", "Luan", "luna?", "Luna?", "luna!", "Luna!"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(message, suffix) {
			return true
		}
	}
	return false
}
