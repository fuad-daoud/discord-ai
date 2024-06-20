package test

//
//import (
//	"flag"
//	"github.com/bwmarrin/discordgo"
//	"log"
//)
//
//var (
//	Token     string
//	ChannelID string
//	GuildID   string
//)
//
//func init() {
//	flag.StringVar(&Token, "t", "", "Bot Token")
//	flag.StringVar(&GuildID, "g", "", "Guild in which voice channel exists")
//	flag.StringVar(&ChannelID, "c", "", "Voice channel to connect to")
//	flag.Parse()
//
//	log.SetFlags(log.Ldate | log.Lmicroseconds)
//}
//func test() {
//
//	//slog.Info("Starting application bot with token", Token)
//	//session, err := discordgo.New("Bot " + Token)
//	//if err != nil {
//	//	slog.Info("error creating Discord session:", err)
//	//	return
//	//}
//	//session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.PermissionViewChannel
//	//session.LogLevel = discordgo.LogDebug
//	//session.AddHandler(BotIsUpHandler())
//	//err = session.Open()
//	//if err != nil {
//	//	log.Fatal("error opening connection:", err)
//	//}
//	//
//	//defer session.Close()
//	//
//	//slog.Info("Bot is now running.  Press CTRL-C to exit.")
//	//stop := make(chan os.Signal, 1)
//	//signal.Notify(stop, os.Interrupt)
//	//<-stop
//	//slog.Info("Graceful shutdown")
//
//}
//
//func BotIsUpHandler() func(s *discordgo.Session, r *discordgo.Ready) {
//	return func(s *discordgo.Session, r *discordgo.Ready) {
//		slog.Info("Bot is up!")
//	}
//}
