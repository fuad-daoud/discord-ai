package deepgram

import (
	"encoding/json"
	"fmt"
	"github.com/forPelevin/gomoji"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	model     = "aura-luna-en"
	encoding  = "opus"
	container = "ogg"
)

func TTS(textToSpeech string) (io.ReadCloser, error) {
	textToSpeech = gomoji.RemoveEmojis("..." + textToSpeech)
	url := fmt.Sprintf("https://api.deepgram.com/v1/speak?model=%s&&encoding=%s&&container=%s", model, encoding, container)
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	payload, _ := json.Marshal(struct {
		Text string `json:"text"`
	}{
		Text: textToSpeech,
	})

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got HTTP status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}
