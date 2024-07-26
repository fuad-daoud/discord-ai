package youtube

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/integrations/digitalocean"
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
	Data          Data
}

func (y *Ytdlp) GetAudio(report func(err error)) (*[][]byte, error) {
	if !y.Data.filled {
		return nil, errors.New("did not search for Data first")
	}
	result, err := y.videoPackets(report)
	if err != nil {
		return nil, err
	}
	segmants := make([][]byte, 0)

	go func() {
		for {
			seg, ok := <-result
			if !ok {
				break
			}
			segmants = append(segmants, seg)
		}
	}()
	return &segmants, nil
}

func (y *Ytdlp) videoPackets(report func(err error)) (chan []byte, error) {
	ytdlpAudio, _ := y.download(report)

	ffmpeg, err := audio.FFMPEG(ytdlpAudio)
	if err != nil {
		return nil, err
	}
	dca := audio.DCA{
		Cache: func(filePath string) {
			y.cache(filePath)
		},
	}
	return dca.Convert(ffmpeg), nil
}

func (y *Ytdlp) download(report func(err error)) (io.Reader, error) {
	start := time.Now()
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
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
		"--no-warnings",
		y.Data.Url,
	)

	dlog.Log.Info("starting youtube command")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
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
						report(err)
						return
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
		return nil, err
	}

	elapsed := time.Since(start)
	dlog.Log.Info("time for ytdlp", "duration", elapsed.Seconds())

	open, err := os.Open("/tmp/audio/" + newUUID.String() + ".opus")
	if err != nil {
		return nil, err
	}

	return open, nil
}

func Search(query string) (Data, error) {
	cmd := exec.CommandContext(context.Background(), "yt-dlp",
		"ytsearch1:"+query,
		"-j",
		"--concurrent-fragments", "16",
		"--audio-format", "opus",
		"--quiet",
		"--no-warnings",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return Data{}, err
	}
	var data Data
	err = json.Unmarshal(output, &data)
	if err != nil {
		//dlog.Log.Warn("could not parse json trying to fetch a new one ...")
		//if _, ok := err.(json.SyntaxError); ok {
		//	time.Sleep(1 * time.Second)
		//	return Search(query)
		//}
		dlog.Log.Error("got an er", "err", err)
		return Data{}, err
	}
	data.filled = true
	return data, nil
}

func (y *Ytdlp) cache(filePath string) {
	tagsJson, _ := json.Marshal(y.Data.Tags)
	categoriesJson, _ := json.Marshal(y.Data.Categories)

	digitalocean.Upload(filePath, "/youtube/cache/"+y.Data.Id+".opus", map[string]*string{
		"Id":         aws.String(y.Data.Id),
		"FullTitle":  aws.String(y.Data.FullTitle),
		"Tags":       aws.String(string(tagsJson)),
		"Categories": aws.String(string(categoriesJson)),
		"ViewCount":  aws.String(strconv.Itoa(y.Data.ViewCount)),
		"Thumbnail":  aws.String(y.Data.Thumbnail),
		//"Description":    aws.String(y.Data.Description),
		"DurationString": aws.String(y.Data.DurationString),
		"LikeCount":      aws.String(strconv.Itoa(y.Data.LikeCount)),
		"Channel":        aws.String(y.Data.Channel),
		"UploaderId":     aws.String(y.Data.UploaderId),
		"Url":            aws.String(y.Data.Url),
	})
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
	filled         bool
}
