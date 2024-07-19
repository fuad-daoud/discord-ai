package youtube

import (
	"github.com/fuad-daoud/discord-ai/integrations/youtube/dca"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Youtube struct {
	Process       func(seg []byte)
	Progress      func(percentage float64)
	ProgressError func(input string)
}

func (y Youtube) Play(id string) {
	data, err := y.VideoPackets(id)
	if err != nil {
		panic(err)
	}
	for {
		seg, ok := <-data
		if !ok {
			break
		}
		go y.Process(seg)
	}
}

func (y Youtube) VideoPackets(id string) (chan []byte, error) {
	outputChannel, err := y.convert(id)
	if err != nil {
		return nil, err
	}
	return outputChannel, nil
}

func (y Youtube) convert(id string) (chan []byte, error) {

	cmd := exec.CommandContext(context.Background(), "ffmpeg",
		"-i",
		"pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)

	cmd.Stdin = y.Download(id)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return dca.DCA(pipe), nil
}

func (y Youtube) Download(id string) io.Reader {
	start := time.Now()

	cmd := exec.CommandContext(context.Background(), "/home/fuad/GolandProjects/discord-ai/artifacts/yt-dlp",
		"--progress-template", "%(progress._percent_str)s",
		"--progress-delta", "1",
		"--no-cache-dir",
		"--no-clean-info-json",
		"--concurrent-fragments", "16",
		"--lazy-playlist",
		"--audio-format", "opus",
		"--no-write-comments",
		"--extract-audio",
		"--quiet",
		//"--color", "no_color",
		//"--no-colors",
		"--progress",
		"--output", "/tmp/audio/"+id,
		//"--simulate",
		"https://www.youtube.com/watch?v="+id,
	)

	dlog.Log.Info("starting youtube command")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go func() {
		buf := make([]byte, 16)
		for {
			n, err2 := stdout.Read(buf)
			if n > 0 {

				input := string(buf[:n])
				if len(input) != 0 {
					input = input[1:]
					input = input[:len(input)-1]
					input = strings.TrimSpace(input)
					if len(input) == 0 {
						continue
					}
					percentage, err := strconv.ParseFloat(input, 64)
					if err != nil {
						panic(err)
					}
					go y.Progress(percentage)
				}
			}
			if err2 != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 16)
		for {
			n, err2 := stderr.Read(buf)
			if n > 0 {
				go y.ProgressError(string(buf[:n]))
			}
			if err2 != nil {
				break
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	dlog.Log.Info("time for ytdlp", "duration", elapsed.Seconds())

	open, err := os.Open("/tmp/audio/" + id + ".opus")
	if err != nil {
		panic(err)
	}

	return open
}
