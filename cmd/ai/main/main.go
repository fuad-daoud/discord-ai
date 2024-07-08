package main

import (
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/http"
	"github.com/fuad-daoud/discord-ai/platform"
	"github.com/fuad-daoud/discord-ai/platform/commands"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	//name, _ := os.Hostname()
	//log.SetPrefix(name)

	//slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	db.Connect()
	go http.Setup()
	go platform.Setup()
	go func() {
		time.Sleep(2 * time.Second)
		commands.AddCommandsChannelOnReadyHandler()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	defer platform.Close()
	defer db.Close()
	slog.Info("Graceful shutdown")
}
