package audio

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

var (
	resultChan = make(chan []byte)
	outputChan = make(chan []byte)
)

func ReadDCA(in io.ReadCloser) (*[][]byte, error) {
	packets := make([][]byte, 0)
	go func() {
		var opuslen int16
		for {
			err := binary.Read(in, binary.LittleEndian, &opuslen)

			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				err := in.Close()
				if err != nil {
					panic(err)
				}
				return
			}
			if err != nil {
				dlog.Log.Error("Error reading from dca file", "err", err)
				panic(err)
			}

			InBuf := make([]byte, opuslen)
			err = binary.Read(in, binary.LittleEndian, &InBuf)

			if err != nil {
				dlog.Log.Error("Error reading from dca file", "err", err)
				panic(err)
			}

			packets = append(packets, InBuf)
		}
	}()
	return &packets, nil
}

func DCA(in io.Reader) chan []byte {
	dlog.Log.Info("starting dca")
	resultChan = make(chan []byte)
	outputChan = make(chan []byte)
	go process(in)
	go write()
	return resultChan
}

func process(in io.Reader) {
	OpusEncoder, err := gopus.NewEncoder(AudioFrameRate, AudioChannels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		panic(err)
	}

	OpusEncoder.SetBitrate(AudioBitrate * 1000)

	OpusEncoder.SetApplication(gopus.Audio)
	defer func() {
		dlog.Log.Info("closing channel ResultChan,OutputChan finished")
		close(resultChan)
		close(outputChan)
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
			resultChan <- bytes
			outputChan <- bytes
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
		resultChan <- bytes
		outputChan <- bytes
	}
}
func write() {
	create, err := os.Create("test.opus")
	if err != nil {
		panic(err)
	}
	stdout := bufio.NewWriterSize(create, 16384)
	defer func() {
		err := stdout.Flush()
		if err != nil {
			dlog.Log.Error("error flushing stdout", "err", err)
		}
	}()
	for {
		bytes, ok := <-outputChan
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
}
func encode(OpusEncoder *gopus.Encoder, buf []int16) ([]byte, error) {
	opus, err := OpusEncoder.Encode(buf, AudioFrameSize, MaxBytes)
	if err != nil {
		fmt.Println("Encoding Error:", err)
		return nil, err
	}
	return opus, nil
}
