package youtube

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/integrations/digitalocean"
	"github.com/fuad-daoud/discord-ai/integrations/youtube/ytclient"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"golang.org/x/net/context"
	"io"
	"os/exec"
	"strconv"
)

type Ytdlp struct {
	Data Data
}

func (y *Ytdlp) GetAudio() (*[][]byte, error) {
	if !y.Data.filled {
		return nil, errors.New("did not search for Data first")
	}
	result, err := y.videoPackets()
	if err != nil {
		return nil, err
	}
	segmants := make([][]byte, 0)

	go func() {
		defer rec()
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
func rec() {
	if r := recover(); r != nil {
		dlog.Log.Error("Recovered ", "msg", r)
	}
}

func (y *Ytdlp) videoPackets() (chan []byte, error) {
	ytdlpAudio, err := y.download()
	if err != nil {
		return nil, err
	}
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

func (y *Ytdlp) download() (io.Reader, error) {
	client := ytclient.Client{}
	video, err := client.GetVideo(y.Data.Id)
	if err != nil {
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return nil, err
	}

	return stream, nil
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

	err := digitalocean.Upload(filePath, "/youtube/cache/"+y.Data.Id+".opus", map[string]*string{
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
		"OriginalUrl":    aws.String(y.Data.Url),
	})
	if err != nil {
		dlog.Log.Error("failed uploading to digital ocean", "err", err)
		return
	}
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
