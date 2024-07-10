package main

import (
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/http"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/fuad-daoud/discord-ai/platform"
	"github.com/fuad-daoud/discord-ai/platform/commands"
	"os"
	"os/signal"
	"time"
)

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
	dlog.Info("Graceful shutdown")
}
