package youtube

import (
	"errors"
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/google/uuid"
	"math"
	"time"
)

type Player interface {
	Run(func(err error))
	Stop()
	Pause()
	Resume()
	Seek(time time.Duration)
	SetConn(conn voice.Conn)
	Add(data Data, packets *[][]byte) error
	Save()
	Skip() error
	GetDBPlayer() DBPlayer
	Idle()
	Clear()
	ToggleLoopQueue() bool
	ToggleLoop() bool
}

type DefaultPlayer struct {
	DBPlayer
	queue     Queue
	inst      InstructionType
	conn      voice.Conn
	GuildId   string
	loopQueue bool
	loop      bool
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

func (p *DefaultPlayer) Add(data Data, packets *[][]byte) error {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return errors.New(err.Error())
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
		OriginalUrl:    data.Url,
		UUID:           newUUID.String(),
	}
	p.addQueueElement(element)
	p.queue.add(&QueueElement{
		DBQueueElement: element,
		Packets:        packets,
	})
	return nil
}

func (p *DefaultPlayer) ReAdd(element *QueueElement) error {
	p.addQueueElement(element.DBQueueElement)
	p.queue.add(element)
	return nil
}

func (p *DefaultPlayer) Run(report func(err error)) {
	if p.inst != IDLE {
		return
	}
	go p.run(report)
}

func (p *DefaultPlayer) run(report func(err error)) {
	dlog.Log.Info("Running player loop")
	if len(p.queue) == 0 {
		dlog.Log.Warn("running with empty queue")
		return
	}
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

				index := element.packetIndex
				if len(seg) != 0 && index >= len(seg) {
					dlog.Log.Info("finished packets")
					if !p.loop {
						poppedElement, err := p.queue.Pop()
						if err != nil {
							report(err)
							return
						}
						if p.loopQueue {
							poppedElement.packetIndex = 0
							err := p.ReAdd(poppedElement)
							if err != nil {
								report(err)
								return
							}
						}
					} else {
						element.packetIndex = 0
					}
					if len(p.queue) > 0 {
						go p.run(report)
					}
					return
				}
				err := p.writeCurrentSeg(seg[index])
				if err != nil {
					report(err)
					return
				}
				time.Sleep(20 * time.Millisecond)
				element.packetIndex++
				break
			}
		case Pause:
			{
				dlog.Log.Info("Got Pause instruction")
				p.inst = IDLE
				return
			}
		case Stop:
			{
				dlog.Log.Info("Got Stop instruction")
				p.Clear()
				p.inst = IDLE
				return
			}
		default:
			dlog.Log.Error("invalid inst", "inst", p.inst)
			return
		}
	}
}
func (p *DefaultPlayer) writeCurrentSeg(seg []byte) error {
	_, err := p.conn.UDP().Write(seg)

	if err != nil {
		dlog.Log.Error("Failed to send talk segment", "error", err)
		return errors.New("failed to send talk segment")
	}
	return nil
}
func (p *DefaultPlayer) Pause() {
	p.inst = Pause
}
func (p *DefaultPlayer) Resume() {
	p.inst = Resume
	go p.run(func(err error) {
		dlog.Log.Error("Failed to resume player", "error", err)
	})
}
func (p *DefaultPlayer) Seek(time time.Duration) {
	dlog.Log.Info("seeking to ", "duration", time)
	seconds := int(time.Seconds())
	p.queue.Head().packetIndex = seconds * 50
}
func (p *DefaultPlayer) Stop() {
	p.inst = Stop
}

func (p *DefaultPlayer) Clear() {
	p.queue.clear()
	p.queue = make(Queue, 0)
}

func (p *DefaultPlayer) Idle() {
	p.inst = IDLE
}

func (p *DefaultPlayer) ToggleLoopQueue() bool {
	if p.loop {
		p.loop = false
	}
	p.loopQueue = !p.loopQueue
	return p.loopQueue
}
func (p *DefaultPlayer) ToggleLoop() bool {
	if p.loopQueue {
		p.loop = false
	}
	p.loop = !p.loop
	return p.loop
}

func (p *DefaultPlayer) Skip() error {
	if len(p.queue) <= 0 {
		dlog.Log.Error("Skipping empty queue")
		return errors.New("skipping empty queue")
	}
	element := p.queue.Head()
	element.packetIndex = math.MaxInt - 1_000
	return nil
}
