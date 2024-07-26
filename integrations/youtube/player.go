package youtube

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/audio"
	"github.com/fuad-daoud/discord-ai/integrations/digitalocean"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/google/uuid"
	"math"
	"time"
)

type Player interface {
	Run()
	Stop()
	Pause()
	Resume()
	Seek(time time.Duration)
	SetConn(conn voice.Conn)
	Add(data Data, packets *[][]byte)
	Save()
	Skip()
	GetDBPlayer() DBPlayer
}

type DefaultPlayer struct {
	DBPlayer
	queue   Queue
	inst    InstructionType
	conn    voice.Conn
	Playing bool
	GuildId string
}

func (p *DefaultPlayer) GetDBPlayer() DBPlayer {
	return p.DBPlayer
}

func (p *DefaultPlayer) Save() {
	p.DBPlayer.Save(p.GuildId)
}

func (p *DefaultPlayer) SetConn(conn voice.Conn) {
	p.conn = conn
}

type Instruction struct {
	Type InstructionType
}

type InstructionType int

const (
	IDLE    InstructionType = iota
	nextSeg InstructionType = iota
	Pause   InstructionType = iota
	Resume  InstructionType = iota
	Stop    InstructionType = iota
	Seek    InstructionType = iota
)

func (p *DefaultPlayer) Add(data Data, packets *[][]byte) {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	element := DBQueueElement{
		Index:          len(p.queue),
		SpaceLink:      "",
		Id:             data.Id,
		FullTitle:      data.FullTitle,
		Tags:           data.Tags,
		Categories:     data.Categories,
		ViewCount:      data.ViewCount,
		Thumbnail:      data.Thumbnail,
		Description:    data.Description,
		DurationString: data.DurationString,
		LikeCount:      data.LikeCount,
		Channel:        data.Channel,
		UploaderId:     data.UploaderId,
		Url:            data.Url,
		UUID:           newUUID.String(),
	}
	p.addQueueElement(element)
	p.queue.add(&QueueElement{
		DBQueueElement: element,
		Packets:        packets,
	})

}

func (p *DefaultPlayer) Run() {
	if p.inst != IDLE {
		return
	}
	go p.run()
}

func (p *DefaultPlayer) run() {
	dlog.Log.Info("Running player loop")

	p.Playing = true
	p.inst = Resume
	element := p.queue.Head()

	for {
		seg := *element.Packets
		if len(seg) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}
		switch p.inst {
		case nextSeg, Resume:
			{
				p.Playing = true

				index := element.packetIndex
				if len(seg) != 0 && index >= len(seg) {
					dlog.Log.Info("finished packets")
					p.queue.Pop()
					if len(p.queue) > 0 {
						go p.run()
					}
					return
				}
				p.writeCurrentSeg(seg[index])
				time.Sleep(20 * time.Millisecond)
				element.packetIndex++
				break
			}
		case Pause:
			{
				dlog.Log.Info("Got Pause instruction")
				p.Playing = false
				return
			}
		case Stop:
			{
				dlog.Log.Info("Got Stop instruction")
				element.packetIndex = 0
				p.Playing = false
				p.clear()
				p.inst = IDLE
				return
			}
		default:
			panic("unhandled default case")
		}
	}
}
func (p *DefaultPlayer) writeCurrentSeg(seg []byte) {
	_, err := p.conn.UDP().Write(seg)

	if err != nil {
		dlog.Log.Error("Failed to send talk segment", "error", err)
		panic(err)
	}
}
func (p *DefaultPlayer) Pause() {
	p.inst = Pause
}
func (p *DefaultPlayer) Resume() {
	p.inst = Resume
	p.Run()
}
func (p *DefaultPlayer) Seek(time time.Duration) {
	dlog.Log.Info("seeking to ", "duration", time)
	seconds := int(time.Seconds())
	p.queue.Head().packetIndex = seconds * 50
}
func (p *DefaultPlayer) Stop() {
	p.inst = Stop
}

func (p *DefaultPlayer) clear() {
	p.queue.clear()
	p.queue = make(Queue, 0)
}

func (p *DefaultPlayer) Skip() {
	if len(p.queue) <= 0 {
		dlog.Log.Error("Skipping empty queue")
		return
	}
	element := p.queue.Head()
	element.packetIndex = math.MaxInt - 1_000
}

type QueueElement struct {
	DBQueueElement
	Packets         *[][]byte
	packetIndex     int
	FinishedLoading *bool
}

type Queue []*QueueElement

func (element *QueueElement) Load() {
	if element.Packets == nil {
		download := digitalocean.Download("youtube/cache/" + element.Id + ".opus")
		if download != nil {
			element.Packets = audio.ReadDCA(download.Body)
		} else {
			y := Ytdlp{
				Progress: func(percentage float64) {
					dlog.Log.Info("downloading", "percentage", percentage)
				},
				ProgressError: func(input string) {
					dlog.Log.Error("download error", "input", input)
				},
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
			element.Packets, err = y.GetAudio()
			if err != nil {
				panic(err)
			}
		}

	}
}

func (q *Queue) Head() *QueueElement {
	return (*q)[0]
}

func (q *Queue) Pop() QueueElement {
	if len(*q) == 0 {
		dlog.Log.Error("popping on empty queue")
		panic("popping on empty queue")
	}
	element := (*q)[0]
	element.Delete()
	*q = (*q)[1:]
	return *element
}

func (q *Queue) add(element *QueueElement) {
	*q = append(*q, element)
}

func (q *Queue) clear() {
	for _, element := range *q {
		element.GoDelete()
	}
}
func (q *Queue) Load() {
	for index := range *q {
		(*q)[index].Load()
	}
}
