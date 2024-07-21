package youtube

import (
	"encoding/json"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Ytdlp struct {
	Progress      func(percentage float64)
	ProgressError func(input string)
}

func (y Ytdlp) GetAudio(link string) (*[][]byte, error) {
	data, err := y.VideoPackets(link)
	if err != nil {
		return nil, err
	}
	segmants := make([][]byte, 0)

	go func() {
		for {
			seg, ok := <-data
			if !ok {
				break
			}
			segmants = append(segmants, seg)
		}
	}()
	return &segmants, nil
}

func (y Ytdlp) VideoPackets(link string) (chan []byte, error) {
	download := y.download(link)
	outputChannel, err := y.convert(download)
	if err != nil {
		return nil, err
	}
	return outputChannel, nil
}

func (y Ytdlp) convert(in io.Reader) (chan []byte, error) {
	ffmpeg, err := audio.FFMPEG(in)
	if err != nil {
		return nil, err
	}

	return audio.DCA(ffmpeg), nil
}

func (y Ytdlp) download(link string) io.Reader {
	start := time.Now()
	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
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
		"--output", "/tmp/audio/"+newUUID.String(),
		//"--simulate",
		link,
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

	open, err := os.Open("/tmp/audio/" + newUUID.String() + ".opus")
	if err != nil {
		panic(err)
	}

	return open
}

func Search(query string) Data {
	cmd := exec.CommandContext(context.Background(), "yt-dlp",
		"ytsearch1:"+query,
		"-j",
		"--concurrent-fragments", "16",
		"--audio-format", "opus",
		"--quiet",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	var data Data
	err = json.Unmarshal(output, &data)
	if err != nil {
		panic(err)
	}
	return data
}

type Data struct {
	Id             string   `json:"id"`
	FullTitle      string   `json:"fulltitle"`
	Tags           []string `json:"tags"`
	Categories     []string `json:"categories"`
	ViewCount      int      `json:"view_count"`
	Thumbnail      string   `json:"thumbnail"`
	Description    string   `json:"description"`
	DurationString string   `json:"duration_string"`
	LikeCount      int      `json:"like_count"`
	Channel        string   `json:"channel"`
	UploaderId     string   `json:"uploader_id"`
	Url            string   `json:"original_url"`
}
