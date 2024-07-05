package main

import (
	"fmt"
	"gopkg.in/hraban/opus.v2"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Open the input MP3 file
	pcmData, err := ioutil.ReadFile("/home/fuad/apps/audio/audio.pcm")
	if err != nil {
		log.Fatalf("Error reading PCM file: %v", err)
	}
	log.Printf("Read %d bytes from input file", len(pcmData))

	// Convert []byte to []int16
	pcmInt16 := byteToInt16(pcmData)
	log.Printf("Converted to %d int16 samples", len(pcmInt16))

	// Set up Opus encoder
	const sampleRate = 24000 // ElevenLabs sample rate
	const channels = 1       // Assuming mono audio
	const application = opus.AppAudio

	encoder, err := opus.NewEncoder(sampleRate, channels, application)
	if err != nil {
		log.Fatalf("Error creating Opus encoder: %v", err)
	}
	log.Printf("Opus encoder created with sample rate %d Hz, %d channels", sampleRate, channels)

	// Encode PCM to Opus
	frameSize := 960      // 40ms frame for 24kHz
	maxPacketSize := 1275 // Maximum size of an Opus packet

	// Open output file
	outFile, err := os.Create("output.opus")
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	frameCount := 0
	totalSamplesEncoded := 0

	for len(pcmInt16) >= frameSize*channels {
		frame := pcmInt16[:frameSize*channels]
		pcmInt16 = pcmInt16[frameSize*channels:]

		packet := make([]byte, maxPacketSize)
		n, err := encoder.Encode(frame, packet)
		if err != nil {
			log.Printf("Error encoding frame %d: %v", frameCount, err)
			continue
		}

		_, err = outFile.Write(packet[:n])
		if err != nil {
			log.Fatalf("Error writing to output file: %v", err)
		}

		frameCount++
		totalSamplesEncoded += frameSize * channels
	}

	log.Printf("Encoded %d frames (%d samples)", frameCount, totalSamplesEncoded)
	log.Printf("Remaining unencoded samples: %d", len(pcmInt16))

	fmt.Println("Conversion complete")
}

func byteToInt16(data []byte) []int16 {
	if len(data)%2 != 0 {
		log.Fatalf("Invalid PCM data length: %d bytes", len(data))
	}
	int16Data := make([]int16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		int16Data[i/2] = int16(uint16(data[i]) | uint16(data[i+1])<<8)
	}
	return int16Data
}
