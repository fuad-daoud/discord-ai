package dca

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"layeh.com/gopus"
	"log"
)

const (
	AudioChannels  = 2
	AudioFrameRate = 48000
	AudioBitrate   = 64
	AudioFrameSize = 960
	MaxBytes       = (AudioFrameSize * AudioChannels) * 2 // max size of opus data
)

func DCA(in io.Reader) chan []byte {

	OpusEncoder, err := gopus.NewEncoder(AudioFrameRate, AudioChannels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		return nil
	}

	OpusEncoder.SetBitrate(AudioBitrate * 1000)

	OpusEncoder.SetApplication(gopus.Audio)

	ResultChan := make(chan []byte)

	go func() {
		var err error
		defer func() {
			close(ResultChan)
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
