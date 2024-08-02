package main

import (
	"github.com/fuad-daoud/discord-ai/http"
	"github.com/fuad-daoud/discord-ai/layers/db"
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
	dlog.Log.Info("Graceful shutdown")
}
func Connect() {
	dbUri := os.Getenv("NEO4J_DATABASE_URL")
	dbUser := os.Getenv("NEO4J_DATABASE_USER")
	dbPassword := os.Getenv("NEO4J_DATABASE_PASSWORD")
	db.Connect(dbUri, dbUser, dbPassword)
}
