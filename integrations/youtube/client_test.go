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
		panic(err)
	}
	youtube := Youtube{
		Process: func(seg []byte) {

		},
		Progress: func(percentage float64) {
			dlog.Log.Info("Progress", "percentage", percentage)
		},
		ProgressError: func(input string) {
			dlog.Log.Error("Something wrong happened", "err", input)
		},
	}
	youtube.Download(id)
	//t.Errorf("got %q, wanted %q", got, want)
}
