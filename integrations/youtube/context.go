package youtube

import (
	"errors"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/integrations/digitalocean"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/google/uuid"
	"strings"
	"time"
)

type QueueElement struct {
	DBQueueElement
	Packets         *[][]byte
	packetIndex     int
	FinishedLoading *bool
}

type Queue []*QueueElement

func (element *QueueElement) Load() error {
	if element.Packets == nil {
		download := digitalocean.Download("youtube/cache/" + element.Id + ".opus")
		if download != nil && *download.ContentLength > 0 {
			element.Packets = audio.ReadDCA(download.Body)
		} else {
			y := Ytdlp{
				Progress: func(percentage float64) {
					dlog.Log.Info("downloading", "percentage", percentage)
				},
				ProgressError: progressError(),
				Data: Data{
					Id:             element.Id,
					FullTitle:      element.FullTitle,
					Tags:           element.Tags,
					Categories:     element.Categories,
					ViewCount:      element.ViewCount,
					Thumbnail:      element.Thumbnail,
					Description:    element.Description,
					DurationString: element.DurationString,
					LikeCount:      element.LikeCount,
					Channel:        element.Channel,
					UploaderId:     element.UploaderId,
					Url:            element.Url,
					filled:         true,
				},
			}
			var err error
			element.Packets, err = y.GetAudio(report())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func progressError() func(input string) {
	builder := strings.Builder{}
	return func(input string) {
		builder.WriteString(input)
		dlog.Log.Error("download error", "input", builder.String())
	}
}

func report() func(err error) {
	return func(err error) {
		newUUID, _ := uuid.NewUUID()
		now := time.Now()
		dlog.Log.Error("reporting problem", "err", err, "uuid", newUUID.String(), "time", now)
	}
}

func (q *Queue) Head() *QueueElement {
	return (*q)[0]
}

func (q *Queue) Pop() (*QueueElement, error) {
	if len(*q) == 0 {
		dlog.Log.Error("popping on empty queue")
		return nil, errors.New("popping on empty queue")
	}
	element := (*q)[0]
	element.Delete()
	*q = (*q)[1:]
	return element, nil
}

func (q *Queue) add(element *QueueElement) {
	*q = append(*q, element)
}

func (q *Queue) clear() {
	for _, element := range *q {
		element.GoDelete()
	}
}
func (q *Queue) Load() error {
	for index := range *q {
		err := (*q)[index].Load()
		if err != nil {
			return err
		}
	}
	return nil
}
