package main

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"github.com/lavalibs/pyro/lavalink"
	"log"
)

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

	lavalink.Connect()

	//log.Println("Starting application bot with token", Token)
	//session, err := discordgo.New("Bot " + Token)
	//if err != nil {
	//	log.Println("error creating Discord session:", err)
	//	return
	//}
	//session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.PermissionViewChannel
	//session.LogLevel = discordgo.LogDebug
	//session.AddHandler(BotIsUpHandler())
	//err = session.Open()
	//if err != nil {
	//	log.Fatal("error opening connection:", err)
	//}
	//
	//defer session.Close()
	//
	//log.Println("Bot is now running.  Press CTRL-C to exit.")
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, os.Interrupt)
	//<-stop
	//log.Println("Graceful shutdown")

}

func BotIsUpHandler() func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	}
}
