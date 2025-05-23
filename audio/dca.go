package audio

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/google/uuid"
	"io"
	"layeh.com/gopus"
	"os"
)

const (
	audioChannels  = 2
	audioFrameRate = 48000
	audioBitrate   = 64
	audioFrameSize = 960
	maxBytes       = (audioFrameSize * audioChannels) * 2 // max size of opus data
)

func ReadDCA(in io.ReadCloser) *[][]byte {
	packets := make([][]byte, 0)
	go func() {
		defer rec()
		var opuslen int16
		for {
			err := binary.Read(in, binary.LittleEndian, &opuslen)

			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				err := in.Close()
				if err != nil {
					dlog.Log.Error("Error closing dca file", "err", err)
					return
				}
				return
			}
			if err != nil {
				dlog.Log.Error("Error reading from dca file", "err", err)
				return
			}

			InBuf := make([]byte, opuslen)
			err = binary.Read(in, binary.LittleEndian, &InBuf)

			if err != nil {
				dlog.Log.Error("Error reading from dca file", "err", err)
				return
			}

			packets = append(packets, InBuf)
		}
	}()
	return &packets
}

type DCA struct {
	resultChan chan []byte
	outputChan chan []byte
	Cache      func(opusFile string)
}

func (d *DCA) Convert(in io.Reader) chan []byte {
	dlog.Log.Info("starting dca")
	d.resultChan = make(chan []byte)
	d.outputChan = make(chan []byte)
	go d.process(in)
	go d.write()
	return d.resultChan
}

func (d *DCA) process(in io.Reader) {
	defer rec()
	OpusEncoder, err := gopus.NewEncoder(audioFrameRate, audioChannels, gopus.Audio)
	if err != nil {
		dlog.Log.Error("Error creating opus encoder", "err", err)
		return
	}

	OpusEncoder.SetBitrate(audioBitrate * 1000)

	OpusEncoder.SetApplication(gopus.Audio)
	defer func() {
		dlog.Log.Info("closing channel ResultChan,OutputChan finished")
		close(d.resultChan)
		close(d.outputChan)
	}()
	stdin := bufio.NewReaderSize(in, 32768)
	for {
		buf := make([]int16, audioFrameSize*audioChannels)
		err = binary.Read(stdin, binary.LittleEndian, &buf)
		if err == io.EOF {
			return
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			opus, err := encode(OpusEncoder, buf)
			if err != nil {
				dlog.Log.Error("Error encoding opus", "err", err)
				return
			}
			d.resultChan <- opus
			d.outputChan <- opus
			return
		}
		if err != nil {
			dlog.Log.Error("Error reading opus", "err", err)
			return
		}
		opus, err := encode(OpusEncoder, buf)
		if err != nil {
			dlog.Log.Error("Failed to encode opus", "err", err)
			return
		}
		d.resultChan <- opus
		d.outputChan <- opus
	}
}
func (d *DCA) write() {
	defer rec()
	newUUID, err := uuid.NewUUID()
	if err != nil {
		dlog.Log.Error("Error generating UUID", "err", err)
		return
	}
	fileName := newUUID.String() + ".opus"
	dirs := "./files/"
	err = os.MkdirAll(dirs, 0777)
	if err != nil {
		dlog.Log.Error("Error creating directory", "err", err)
		return
	}
	filePath := dirs + fileName
	create, err := os.Create(filePath)
	if err != nil {
		dlog.Log.Error("Error creating file", "err", err)
		return
	}
	buffer := bufio.NewWriter(create)
	defer func() {
		err := buffer.Flush()
		if err != nil {
			dlog.Log.Error("error flushing stdout", "err", err)
			return
		}
		err = create.Close()
		if err != nil {
			dlog.Log.Error("error flushing stdout", "err", err)
			return
		}
		d.Cache(filePath)
	}()
	for {
		opus, ok := <-d.outputChan
		if !ok {
			break
		}
		opuslen := int16(len(opus))
		err := binary.Write(buffer, binary.LittleEndian, &opuslen)
		if err != nil {
			dlog.Log.Error("error writing output", "err", err)
			return
		}
		err = binary.Write(buffer, binary.LittleEndian, &opus)
		if err != nil {
			dlog.Log.Error("error writing output", "err", err)
			return
		}
	}
}

func encode(OpusEncoder *gopus.Encoder, buf []int16) ([]byte, error) {
	opus, err := OpusEncoder.Encode(buf, audioFrameSize, maxBytes)
	if err != nil {
		fmt.Println("Encoding Error:", err)
		return nil, err
	}
	return opus, nil
}

func rec() {
	if r := recover(); r != nil {
		dlog.Log.Error("Recovered ", "msg", r)
	}
}
