package dca

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"io"
	"layeh.com/gopus"
	"log"
	"os"
)

const (
	AudioChannels  = 2
	AudioFrameRate = 48000
	AudioBitrate   = 64
	AudioFrameSize = 960
	MaxBytes       = (AudioFrameSize * AudioChannels) * 2 // max size of opus data
)

func DCA(in io.Reader) chan []byte {
	dlog.Log.Info("starting dca")
	OpusEncoder, err := gopus.NewEncoder(AudioFrameRate, AudioChannels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		return nil
	}

	OpusEncoder.SetBitrate(AudioBitrate * 1000)

	OpusEncoder.SetApplication(gopus.Audio)

	ResultChan := make(chan []byte)
	OutputChan := make(chan []byte)

	go func() {
		var err error
		defer func() {
			dlog.Log.Info("closing channel ResultChan,OutputChan finished")
			close(ResultChan)
			close(OutputChan)
		}()
		stdin := bufio.NewReaderSize(in, 32768)
		for {
			buf := make([]int16, AudioFrameSize*AudioChannels)
			err = binary.Read(stdin, binary.LittleEndian, &buf)
			if err == io.EOF {
				return
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				bytes, err := encode(OpusEncoder, buf)
				if err != nil {
					return
				}
				ResultChan <- bytes
				OutputChan <- bytes
				return
			}
			if err != nil {
				log.Println("error reading from stdin,", err)
				return
			}
			bytes, err := encode(OpusEncoder, buf)
			if err != nil {
				return
			}
			ResultChan <- bytes
			OutputChan <- bytes
		}
	}()
	go func() {
		create, err := os.Create("test.opus")
		if err != nil {
			panic(err)
		}
		stdout := bufio.NewWriterSize(create, 16384)
		defer func() {
			err := stdout.Flush()
			if err != nil {
				log.Println("error flushing stdout, ", err)
			}
		}()
		for {
			bytes, ok := <-OutputChan
			if !ok {
				break
			}
			opuslen := int16(len(bytes))
			err = binary.Write(stdout, binary.LittleEndian, &opuslen)
			if err != nil {
				dlog.Log.Error("error writing output", "err", err)
				return
			}
			err = binary.Write(stdout, binary.LittleEndian, &bytes)
			if err != nil {
				dlog.Log.Error("error writing output", "err", err)
				return
			}
		}
	}()
	return ResultChan
}

func encode(OpusEncoder *gopus.Encoder, buf []int16) ([]byte, error) {
	opus, err := OpusEncoder.Encode(buf, AudioFrameSize, MaxBytes)
	if err != nil {
		fmt.Println("Encoding Error:", err)
		return nil, err
	}
	return opus, nil
}
