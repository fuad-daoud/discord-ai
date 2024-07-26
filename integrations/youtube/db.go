package youtube

import (
	"fmt"
	"github.com/fuad-daoud/discord-ai/db"
	"github.com/fuad-daoud/discord-ai/db/cypher"
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
	Url            string   `json:"original_url"`
	UUID           string   `json:"UUID"`
	filled         bool
}

func (p DBPlayer) Save(guildId string) {
	g := db.Guild{Id: guildId}
	db.InTransaction(func(write db.Write) {
		write(cypher.MatchN("g", g), cypher.MergeN("p", p), cypher.Merge("(g)-[:HAS]->(p)"))
	})
}

func (p DBPlayer) addQueueElement(q DBQueueElement) {
	db.InTransaction(func(write db.Write) {
		write(cypher.MatchN("p", p), cypher.CreateN("q", q), cypher.Create("(p)-[:QUEUE]->(q)"))
	})
}

func (q DBQueueElement) Delete() {
	db.InTransaction(func(write db.Write) {
		write(cypher.Match(fmt.Sprintf("(q:DBQueueElement {UUID: \"%s\"})", q.UUID)), "DETACH", cypher.Delete("q"))
	})
}
func (q DBQueueElement) GoDelete() {
	go q.Delete()
}

func GetQueue(p DBPlayer) Queue {
	result := db.Query(cypher.MatchN("p", p), "-[:QUEUE]->", "(q)", cypher.Return("q"))
	all, ok := cypher.ParseAll[DBQueueElement]("q", result)
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
