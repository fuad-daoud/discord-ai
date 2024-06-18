package logic

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
