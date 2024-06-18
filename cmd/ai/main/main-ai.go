package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/discordgo/examples/voice_receive/deepgram"
	"github.com/bwmarrin/discordgo/examples/voice_receive/gpt"
	"github.com/bwmarrin/discordgo/examples/voice_receive/logic"
	"github.com/bwmarrin/discordgo/examples/voice_receive/respeecher"
	"log"
	"os"
	"os/signal"
	"time"
)

// Variables used for command line parameters
var (
	Token     string
	ChannelID string
	GuildID   string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "Guild in which voice channel exists")
	flag.StringVar(&ChannelID, "c", "", "Voice channel to connect to")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func main() {
	log.Println("Starting application bot with token", Token)
	session, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("error creating Discord session:", err)
		return
	}

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.PermissionViewChannel

	session.LogLevel = discordgo.LogDebug

	var deepgramClient deepgram.Client = &deepgram.DefaultClient{}

	gptClient := gpt.MakeClient()
	gptClient.CreateThread()

	chanRun := gptClient.GetChanRequiredAction()
	respeecherClient := respeecher.MakeClient()

	go func() {
		for run := range chanRun {
			if run.Status != "requires_action" {
				continue
			}
			functionName := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Function.Name
			if functionName == "join" {
				log.Println("starting join function")
				toolCallId := run.RequiredAction.SubmitToolOutputs.ToolCalls[0].Id
				data := run.MetaData
				if err != nil {
					gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
						Success:     "false",
						Description: err.Error(),
					})
					return
				}

				state, err := session.State.VoiceState(GuildID, data.UserId)
				if err != nil {
					gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
						Success:     "false",
						Description: "User is not in a voice channel please join a voice channel first",
					})
					return
				}
				channelId := state.ChannelID

				if channelId == "" {
					gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
						Success:     "false",
						Description: "User is not in a voice channel please join a voice channel first",
					})
					return
				}
				voiceConnection, err := session.ChannelVoiceJoin(GuildID, channelId, false, false)
				if err != nil {
					gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
						Success:     "false",
						Description: err.Error(),
					})
					return
				}
				go func() {
					go addDeepgramVoiceHandler(voiceConnection, deepgramClient, session, finishedCallBack(voiceConnection, session, gptClient, respeecherClient, data.ChannelId))
					voiceConnection.AddHandler(VoiceSpeakingUpdateHandler(deepgramClient))

					err = logic.Talk(voiceConnection, "files/fixed-replies/hey-luna-is-here-oksana.wav")
					if err != nil {
						log.Println("error talking to voice channel:", err)
						return
					}

				}()

				newRun := gptClient.SubmitToolOutputs(toolCallId, gpt.OutputTool{
					Success:     "true",
					Description: "",
				})

				chanRun <- *newRun
			}
		}
	}()

	session.AddHandler(BotIsUpHandler())
	session.AddHandler(VoiceStateUpdateHandler(deepgramClient))
	session.AddHandler(func(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
		if messageCreate.Author.ID != session.State.User.ID {
			channel, err := session.Channel(messageCreate.ChannelID)
			if err != nil {
				log.Println("error getting channel:", err)
			}
			data := gpt.MetaData{
				UserId:    messageCreate.Author.ID,
				ChannelId: channel.ID,
			}

			if channel.IsThread() {
				var process Process
				if channel.ParentID == "1252536839886082109" {
					process = func(message string, data gpt.MetaData) string {
						response := gptClient.SendMessageFullCycle(message+"(Whispering)", data)
						go func() {
							state, err := session.State.VoiceState(GuildID, session.State.User.ID)
							if err != nil {
								return
							}
							if state.ChannelID == "" {
								return
							}
							speech, err := respeecherClient.TextToSpeech(response, respeecher.VoiceParams{
								Id:     respeecher.OksanaDefault.Id,
								Accent: respeecher.OksanaDefault.Accent,
								Style:  respeecher.Oksana.Styles["HushedRaspy"],
							})
							if err != nil {
								return
							}

							connections := session.VoiceConnections
							for _, connection := range connections {
								if connection.UserID == session.State.User.ID {
									logic.Talk(connection, speech)
								}

							}
						}()
						return response
					}
				} else {
					process = gptClient.SendMessageFullCycle
				}
				replyText(data, messageCreate.Message.Content, session, process)
			} else {
				start, err := session.MessageThreadStart(channel.ID, messageCreate.ID, gptClient.GetThreadId(), 1440)
				if err != nil {
					log.Fatal("could not create thread", err.Error())
				}
				log.Println(start.ID)
				data.ChannelId = start.ID
				replyText(data, messageCreate.Message.Content, session, gptClient.SendMessageFullCycle)

			}
		}
	})
	err = session.Open()
	if err != nil {
		log.Fatal("error opening connection:", err)
	}

	defer session.Close()

	//start, err := session.ThreadStart("1252273230727876619", gptClient.GetThreadId(), discordgo.ChannelTypeGuildPublicThread, 1440)
	//if err != nil {
	//	log.Fatal("could not create thread", err.Error())
	//}
	//log.Println(start.ID)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func VoiceSpeakingUpdateHandler(deepgramClient deepgram.Client) func(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
	return func(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
		deepgramClient.MapSSRC(vs.SSRC, vs.UserID)
		log.Println("checkWhoIsTalkingHandler here")
	}
}

func VoiceStateUpdateHandler(deepgramClient deepgram.Client) func(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
		if event.UserID == s.State.User.ID {
			log.Println("Update on bot voice state")
			if len(event.ChannelID) == 0 {
				log.Println("Disconnected from voice channel")
				deepgramClient.Stop()
			}
		} else {

		}
	}
}

func BotIsUpHandler() func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	}
}

func finishedCallBack(voiceConnection *discordgo.VoiceConnection, session *discordgo.Session, gptClient gpt.Client, respeecherClient respeecher.Client, channelId string) deepgram.FinishedCallBack {
	return func(message string, userId string) {
		go handleThread(session, channelId, userId, message)

		voiceConnection, _ = session.ChannelVoiceJoin(voiceConnection.GuildID, voiceConnection.ChannelID, false, true)

		//go indicator(voiceConnection, "processing-this-wont-take-long.wav")
		go indicator(voiceConnection, "rizz-sounds.mp3")
		//go indicator(voiceConnection, "formula-1-radio-notification.mp3")

		botId := session.State.User.ID

		//text := "Hey there, I'm Luna, your stunning Discord bot. What can I do for you today?"
		//text := "I don't have access to real-time information like the current date. But you can check on your device, love. Is there anything else you'd like me to do?"

		response := gptClient.SendMessageFullCycle(message, gpt.MetaData{
			UserId:    userId,
			ChannelId: channelId,
		})

		path, err := respeecherClient.DefaultTextToSpeech(response)
		if err != nil {
			log.Fatal(err)
		}
		go handleThread(session, channelId, botId, response)
		err = logic.Talk(voiceConnection, path)
		if err != nil {
			log.Fatal(err)
		}
		voiceConnection, _ = session.ChannelVoiceJoin(voiceConnection.GuildID, voiceConnection.ChannelID, false, false)
	}
}

func replyText(data gpt.MetaData, message string, session *discordgo.Session, process Process) {
	processingMessage := fmt.Sprintf("%s", "Dazzling✨💫")
	send, err := session.ChannelMessageSend(data.ChannelId, processingMessage)
	if err != nil {
		log.Fatal(err)
	}

	response := process(message, data)

	edit, err := session.ChannelMessageEdit(data.ChannelId, send.ID, response)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("editID: %#v\n", edit.ID)
}

type Process func(message string, data gpt.MetaData) string

func handleThread(session *discordgo.Session, threadId string, userId string, message string) {
	err := session.ThreadMemberAdd(threadId, userId)
	if err != nil {
		log.Fatal(err)
	}
	user, err := session.User(userId)
	if err != nil {
		log.Fatal(err)
	}
	message = fmt.Sprintf("%s: %s", user.Mention(), message)
	send, err := session.ChannelMessageSend(threadId, message)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(send.ID)
}

func addDeepgramVoiceHandler(voiceConnection *discordgo.VoiceConnection, deepgram deepgram.Client, session *discordgo.Session, finishedCallback deepgram.FinishedCallBack) {
	for p := range voiceConnection.OpusRecv {
		state, err := session.State.VoiceState(GuildID, session.State.User.ID)
		if err != nil {
			log.Fatal(err)
		}
		if state.SelfDeaf {
			continue
		}
		deepgram.Write(p.Opus, p.SSRC, finishedCallback)
	}
}

func indicator(voiceConnection *discordgo.VoiceConnection, file string) {
	time.Sleep(1 * time.Second)
	err := logic.Talk(voiceConnection, "files/fixed-replies/"+file)
	if err != nil {
		log.Fatal(err)
	}
}
