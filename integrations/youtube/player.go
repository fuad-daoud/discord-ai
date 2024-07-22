package youtube

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"time"
)

type Player interface {
	Run()
	Stop()
	Pause()
	Resume()
	Seek(time time.Duration)
}

type DefaultPlayer struct {
	Segments     *[][]byte
	currentIndex int
	inst         InstructionType
	Conn         voice.Conn
	Playing      bool
}

type Instruction struct {
	Type InstructionType
}

type InstructionType int

const (
	nextSeg InstructionType = iota
	Pause   InstructionType = iota
	Resume  InstructionType = iota
	Stop    InstructionType = iota
	Seek    InstructionType = iota
)

func (p *DefaultPlayer) Run() {
	p.Playing = true
	go p.run()
}

func (p *DefaultPlayer) run() {
	dlog.Log.Info("Running player loop")

	p.inst = Resume
	for {
		seg := *p.Segments
		if len(seg) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}
		switch p.inst {
		case nextSeg, Resume:
			{
				if p.currentIndex >= len(seg) {
					dlog.Log.Info("Finished segments")
					break
				}
				p.writeCurrentSeg()
				time.Sleep(20 * time.Millisecond)
				p.currentIndex++
				break
			}
		case Pause:
			{
				dlog.Log.Info("Got Pause instruction")
				p.Playing = true
				return
			}
		case Stop:
			{
				dlog.Log.Info("Got Stop instruction")
				p.currentIndex = 0
				p.Playing = true
				return
			}

		}
	}
}

func (p *DefaultPlayer) writeCurrentSeg() {
	seg := *p.Segments
	_, err := p.Conn.UDP().Write(seg[p.currentIndex])

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
	p.currentIndex = seconds * 50
}
func (p *DefaultPlayer) Stop() {
	p.inst = Stop
}
