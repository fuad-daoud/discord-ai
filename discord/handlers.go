package discord

import (
	"fmt"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/fuad-daoud/discord-ai/integrations/deepgram"
	"github.com/fuad-daoud/discord-ai/integrations/gpt"
	"github.com/fuad-daoud/discord-ai/integrations/respeecher"
	"golang.org/x/net/context"
	"log"
	"time"
)

func handleDeepgramVoicePackets(conn voice.Conn, deepgram deepgram.Client, finishedCallback deepgram.FinishedCallBack, client bot.Client) {

	for {
		packet, err := conn.UDP().ReadPacket()
		if err != nil {
			return
		}
		guildId := snowflake.MustParse(`847908927554322432`)

		voiceState, b := client.Caches().VoiceState(guildId, client.ApplicationID())
		if !b {
			log.Fatal("bot not connected to a voice channel")
		}
		if voiceState.SelfDeaf {
			continue
		}
		deepgram.Write(packet.Opus, packet.SSRC, finishedCallback)
	}
}

func finishedCallBack(conn voice.Conn, client bot.Client, gptClient gpt.Client, respeecherClient respeecher.Client, channelId string) deepgram.FinishedCallBack {
	return func(message string, SSRC uint32) {
		log.Println("got SSRC", SSRC)
		userId := conn.UserIDBySSRC(SSRC)
		user, err := client.Rest().GetUser(userId)
		if err != nil {
			log.Fatal("could not get user: ", userId, err)
		}
		detect, response := gptClient.Detect(message, gpt.MetaData{
			UserId:    userId.String(),
			ChannelId: "",
		})
		if !detect {
			return
		}
		go handleThread(client, channelId, *user, message)

		guildId := snowflake.MustParse(`847908927554322432`)

		userState, userStateOk := client.Caches().VoiceState(guildId, userId)
		if !userStateOk {
			log.Fatal("User voice state not okay")
		}
		err = client.UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, true)
		if err != nil {
			log.Fatal("could not update voice state: ", err)
		}

		path, err := respeecherClient.DefaultTextToSpeech(response)
		if err != nil {
			log.Fatal(err)
		}
		selfUser, b := client.Caches().SelfUser()
		if !b {
			log.Fatal("could not get self user")
		}
		go handleThread(client, channelId, selfUser.User, response)

		err = Talk(conn, path, func() error {
			return nil
		}, func() error {
			return client.UpdateVoiceState(context.Background(), guildId, userState.ChannelID, false, false)
		})
		if err != nil {
			log.Fatal(err)
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
//			log.Fatal(err)
//		}
//	}
func handleThread(client bot.Client, threadId string, user discord.User, message string) {
	message = fmt.Sprintf("%s: %s", user.Mention(), message)
	createMessage, err := client.Rest().CreateMessage(snowflake.MustParse(threadId), discord.MessageCreate{Content: message})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created message ", createMessage.ID)
}
func CommandsHandler(client bot.Client, gptClient gpt.Client, deepgramClient deepgram.Client, respeecherClient respeecher.Client) {
	chanRun := gptClient.GetChanRequiredAction()
	go CommandsProcessor(chanRun, client, gptClient, deepgramClient, respeecherClient)
}

func MessageCreateHandler(deepgramClient deepgram.Client, respeecherClient respeecher.Client) func(event *events.MessageCreate) {
	return func(event *events.MessageCreate) {
		if event.Message.Author.ID == event.Client().ID() {
			return
		}
		restClient := event.Client().Rest()
		channel, err := restClient.GetChannel(event.ChannelID)
		if err != nil {
			return
		}
		log.Println("Got channel ", channel.Name())

		data := gpt.MetaData{
			UserId:    event.Message.Author.ID.String(),
			ChannelId: channel.ID().String(),
		}

		if channel.Type() == discord.ChannelTypeGuildPublicThread {
			var process Process
			thread := channel.(discord.GuildThread)
			log.Println("Got thread ", thread.ID())
			if thread.ParentID().String() == "1252536839886082109" {
				process = func(message string, data gpt.MetaData) string {
					botState, botStateOk := event.Client().Caches().VoiceState(*event.GuildID, event.Client().ApplicationID())

					userState, userStateOk := event.Client().Caches().VoiceState(*event.GuildID, event.Message.Author.ID)
					if !userStateOk {
						return "You are not in a voice channel bro "
					}
					go func() {
						err := event.Client().Rest().SendTyping(thread.ID())
						if err != nil {
							log.Fatal(err)
						}
					}()
					go func() {
						if botStateOk && botState.ChannelID.String() == userState.ChannelID.String() {

							if err != nil {
								log.Fatal(err)
							}
							err := event.Client().UpdateVoiceState(context.Background(), *event.GuildID, userState.ChannelID, false, true)
							if err != nil {
								log.Fatal(err)
							}
						}
					}()
					gptClient := *gpt.GetClient(thread.Name())
					go CommandsHandler(event.Client(), gptClient, deepgramClient, respeecherClient)
					response := gptClient.SendMessageFullCycle(message+"(respond like you are whispering)", data)

					speech, err := respeecherClient.TextToSpeech(response, respeecher.VoiceParams{
						Id:     respeecher.OksanaDefault.Id,
						Accent: respeecher.OksanaDefault.Accent,
						Style:  respeecher.Oksana.Styles["HushedRaspy"],
					})
					if err != nil {
						log.Fatal(err)
					}
					go func() {
						if botStateOk && botState.ChannelID.String() == userState.ChannelID.String() {
							conn := event.Client().VoiceManager().GetConn(*event.GuildID)
							err = Talk(conn, speech, func() error {
								return event.Client().UpdateVoiceState(context.Background(), *event.GuildID, userState.ChannelID, false, true)
							}, func() error {
								return event.Client().UpdateVoiceState(context.Background(), *event.GuildID, userState.ChannelID, false, false)
							})
							if err != nil {
								log.Fatal(err)
							}
							return
						}
						ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
						defer cancel()
						conn := event.Client().VoiceManager().CreateConn(*event.GuildID)
						err = conn.Open(ctx, *userState.ChannelID, false, false)
						if err != nil {
							log.Fatal(err)
						}
						err = Talk(conn, speech, func() error {
							return event.Client().UpdateVoiceState(context.Background(), *event.GuildID, userState.ChannelID, false, true)
						}, func() error {
							return event.Client().UpdateVoiceState(context.Background(), *event.GuildID, userState.ChannelID, false, false)
						})
						if err != nil {
							log.Fatal(err)
						}
					}()
					return response
				}
			} else {
				gptClient := *gpt.GetClient(thread.Name())
				go CommandsHandler(event.Client(), gptClient, deepgramClient, respeecherClient)
				process = gptClient.SendMessageFullCycle
			}
			replyText(data, event.Message.Content, restClient, process)
		} else {
			if channel.ID().String() != "1252273230727876619" && channel.ID().String() != "1252536839886082109" {
				return
			}
			gptClient := gpt.MakeClient()
			gptThreadId := gptClient.GetThreadId()
			go CommandsHandler(event.Client(), gptClient, deepgramClient, respeecherClient)
			detect, response := gptClient.Detect(event.Message.Content, data)
			if !detect {
				return
			}
			thread, err := restClient.CreateThreadFromMessage(channel.ID(), event.MessageID, discord.ThreadCreateFromMessage{Name: gptThreadId, AutoArchiveDuration: 1440})
			if err != nil {
				log.Fatal("could not create thread", err.Error())
			}
			log.Println("Created thread id", thread.ID())
			data.ChannelId = thread.ID().String()
			replyText(data, event.Message.Content, restClient, func(message string, data gpt.MetaData) string {
				return response
			})
		}

	}
}

func BotIsUp(r *events.Ready) {
	log.Println("Bot is up!")
}
func replyText(data gpt.MetaData, content string, client rest.Rest, process Process) {
	processingMessage := fmt.Sprintf("%s", "Dazzling✨💫")

	channelId := snowflake.MustParse(data.ChannelId)
	message, err := client.CreateMessage(channelId, discord.MessageCreate{
		Content: processingMessage,
	})
	if err != nil {
		log.Fatal(err)
	}

	response := process(content, data)
	updateMessage, err := client.UpdateMessage(channelId, message.ID, discord.MessageUpdate{Content: &response})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("updated message: %#v\n", updateMessage.ID.String())
}

type Process func(message string, data gpt.MetaData) string

func VoiceServerUpdateHandler(deepgramClient deepgram.Client) func(event *events.GuildVoiceStateUpdate) {
	return func(event *events.GuildVoiceStateUpdate) {

		if event.Member.User.ID == event.Client().ID() {
			log.Println("Update on bot voice state")

			newChannelId := event.GenericGuildVoiceState.VoiceState.ChannelID.String()

			if len(newChannelId) == 0 {
				log.Println("Disconnected from voice channel")
				deepgramClient.Stop()
			}
		} else {

		}
	}
}
