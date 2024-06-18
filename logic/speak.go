package logic

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"sync"
)

var mutex = sync.Mutex{}

func Talk(connection *discordgo.VoiceConnection, filePath string) error {
	sound, err := LoadSound(filePath)
	if err != nil {
		return err
	}
	err = PlaySound(connection, sound)
	return err
}

func PlaySound(voiceConnection *discordgo.VoiceConnection, buffer [][]byte) (err error) {

	mutex.Lock()

	_ = voiceConnection.Speaking(true)

	for _, buff := range buffer {
		voiceConnection.OpusSend <- buff
	}

	_ = voiceConnection.Speaking(false)
	mutex.Unlock()
	return nil
}

func LoadSound(filePath string) ([][]byte, error) {

	dca, err := convertMp3ToDca(filePath)
	if err != nil {
		panic(err)
	}

	var buffer = make([][]byte, 0)

	file, err := os.Open(dca)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return nil, err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return nil, err
			}
			break
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
	return buffer, nil
}
