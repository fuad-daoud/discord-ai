package deepgram

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const (
	filePath string = "files/output.mp3"
	model           = "aura-luna-en"
	encoding        = "mp3"
)

func TTS(textToSpeech string) (string, error) {
	url := fmt.Sprintf("https://api.deepgram.com/v1/speak?model=%s&&encoding=%s", model, encoding)
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	payload := strings.NewReader(`{
				"text": "` + textToSpeech + `"
				}`)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	outputFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		return "", err

	}
	slog.Info("File saved successfully.")
	return filePath, nil
}
