package youtube

import (
	"fmt"
	db2 "github.com/fuad-daoud/discord-ai/layers/db"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"sort"
)

type DBPlayer struct {
	Id     string
	Status string
}

type DBQueueElement struct {
	Index          int      `json:"index"`
	SpaceLink      string   `json:"spaceLink"`
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
	OriginalUrl    string   `json:"original_url"`
	UUID           string   `json:"UUID"`
	filled         bool
}

func (p DBPlayer) Save(guildId string) {
	g := db2.Guild{Id: guildId}
	err := db2.Transaction(func(write db2.Write) error {
		err := write(db2.MatchN("g", g), db2.MergeN("p", p), db2.Merge("(g)-[:HAS]->(p)"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		dlog.Log.Error("Failed to insert player in DB", guildId, err)
	}
}

func (p DBPlayer) addQueueElement(q DBQueueElement) {
	err := db2.Transaction(func(write db2.Write) error {
		return write(db2.MatchN("p", p), db2.CreateN("q", q), db2.Create("(p)-[:QUEUE]->(q)"))
	})
	if err != nil {
		dlog.Log.Error("error adding queue element: ", err)
	}
}

func (q DBQueueElement) Delete() {
	err := db2.Transaction(func(write db2.Write) error {
		return write(db2.Match(fmt.Sprintf("(q:DBQueueElement {UUID: \"%s\"})", q.UUID)), "DETACH", db2.Delete("q"))
	})
	if err != nil {
		dlog.Log.Error("error deleting queue element: ", err)
	}
}
func (q DBQueueElement) GoDelete() {
	go q.Delete()
}

func GetQueue(p DBPlayer) Queue {
	result, err := db2.Query(db2.MatchN("p", p), "-[:QUEUE]->", "(q)", db2.Return("q"))
	if err != nil {
		dlog.Log.Error("error getting queue element: ", err)
		//TODO: remove this panic
		panic("error getting queue element")
	}
	all, ok := db2.ParseAll[DBQueueElement]("q", result.Records)
	if !ok {
		return Queue{}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Index < all[j].Index
	})
	q := make(Queue, len(all))
	for index, element := range all {
		queueElement := QueueElement{
			DBQueueElement: element,
		}
		q[index] = &queueElement
	}
	return q
}
