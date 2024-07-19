package main

import (
	"context"
	"fmt"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)

	dl := ytdlp.New().
		FormatSort("ext:opus").
		ExtractAudio().
		NoCacheDir().
		CleanInfoJSON().
		ConcurrentFragments(16).
		Format("bestaudio/best").
		AudioFormat("opus").
		NoWriteComments().
		LazyPlaylist().
		Output("test.%(ext)s")

	Result, err := dl.Run(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		panic(err)
	}
	fmt.Print(Result)
}

//https://github.com/iawia002/lux
//https://github.com/wader/goutubedl
