package discord

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/disgoorg/disgo/voice"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var mutex = sync.Mutex{}

func Talk(conn voice.Conn, filePath string, def func() error, undef func() error) error {
	sound, err := LoadSound(filePath)
	if err != nil {
		return err
	}
	err = PlaySound(conn, sound, def, undef)
	return err
}

func PlaySound(conn voice.Conn, buffer [][]byte, def func() error, undef func() error) (err error) {
	mutex.Lock()
	err = def()
	if err != nil {
		return err
	}
	err = conn.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone)
	if err != nil {
		return err
	}
	if _, err := conn.UDP().Write(voice.SilenceAudioFrame); err != nil {
		return err
	}
	log.Println("Starting writing packets")
	for _, buff := range buffer {
		_, err := conn.UDP().Write(buff)
		if err != nil {
			log.Fatal()
		}
		time.Sleep(20 * time.Millisecond)
	}
	log.Println("Finished writing packets")
	mutex.Unlock()
	err = undef()
	if err != nil {
		return err
	}
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

func convertMp3ToDca(mp3filePath string) (string, error) {

	if _, err := os.Stat(mp3filePath); os.IsNotExist(err) {
		log.Printf("mp3 file not exists: %s", mp3filePath)
		panic(err)
	}

	dcaFile := changeDirAndExtension(mp3filePath)

	if _, err := os.Stat(dcaFile); !os.IsNotExist(err) {
		log.Printf("Already converted mp3 file %s", dcaFile)
		return dcaFile, nil
	}

	ffmpegCmd := exec.Command("ffmpeg",
		"-i", mp3filePath,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)
	// Define the dca command and its output file
	dcaCmd := exec.Command("/opt/dca")
	outputFile := changeDirAndExtension(mp3filePath)

	// Pipe the output of ffmpeg to dca
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	dcaCmd.Stdin = ffmpegOut
	// Create the output file
	dcaOut, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dcaOut.Close()
	dcaCmd.Stdout = dcaOut

	// Start the ffmpeg command
	if err := ffmpegCmd.Start(); err != nil {
		log.Fatal(err)
		return "", err
	}

	// Start the dca command
	if err := dcaCmd.Start(); err != nil {
		log.Fatal(err)
		return "", err
	}

	// Wait for ffmpeg to finish
	if err := ffmpegCmd.Wait(); err != nil {
		log.Fatal(err)
		return "", err
	}

	// Wait for dca to finish
	if err := dcaCmd.Wait(); err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Println("Conversion completed successfully!")
	log.Println("dca filepath :", outputFile)
	return outputFile, nil
}

func changeDirAndExtension(filePath string) string {
	// Change the file extension from .mp3 to .dca
	newFilePath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".dca"

	// Change the parent directory from "mp3" to "dca"
	newFilePath = strings.Replace(newFilePath, "/mp3/", "/dca/", 1)

	return newFilePath
}
