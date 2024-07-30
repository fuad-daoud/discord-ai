package ytclient_test

import (
	"fmt"
	youtube "github.com/fuad-daoud/discord-ai/integrations/youtube/ytclient"
	"io"
	"os"
	"testing"
	"time"
)

// ExampleDownload : Example code for how to use this package for download video.
func TestExampleClient(t *testing.T) {
	start := time.Now()
	videoID := "-ybCiHPWKNA"
	client := youtube.Client{}

	video, err := client.GetVideo(videoID)
	if err != nil {
		panic(err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	file, err := os.Create("video.mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}
	since := time.Since(start)
	fmt.Println(since)
}
