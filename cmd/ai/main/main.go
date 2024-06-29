package main

import (
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/http"
	"github.com/fuad-daoud/discord-ai/platform"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	//slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	db.Connect()
	go platform.CancelAllPendingRuns()
	go http.Setup()
	go platform.Setup()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	defer platform.Close()
	defer db.Close()
	slog.Info("Graceful shutdown")
}
