package youtube

import (
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"os"
	"testing"
)

func TestYoutubeProgress(t *testing.T) {
	id := "LriHRa9t1fQ"
	//id := "_QLZs0QTaZM"
	//id := "91wX0NRjJqg"
	//id := ""Vxq6Qc-uAmE"

	err := os.RemoveAll("/tmp/audio/" + id + ".opus")
	if err != nil {
		os.Exit(1)
	}
	youtube := Ytdlp{
		Progress: func(percentage float64) {
			dlog.Log.Info("Progress", "percentage", percentage)
		},
		ProgressError: func(input string) {
			dlog.Log.Error("Something wrong happened", "err", input)
		},
	}
	youtube.Data = Search("LriHRa9t1fQ")
	youtube.download()
	//t.Errorf("got %q, wanted %q", got, want)
}

func TestYoutubeSearch(t *testing.T) {
	search := Search("cute cats")
	dlog.Log.Info("output of search", "search", search)
}
