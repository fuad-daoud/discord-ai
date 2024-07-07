package elevenlabs

import (
	"context"
	"errors"
	"fmt"
	"github.com/fuad-daoud/discord-ai/integrations/elevenlabs/ffmpeg"
	"net/http"
	"os"
	"strings"
)

var (
	stability  = 0.30
	similarity = 0.78
	//voiceId    = "eVItLK1UvXctxuaRV2Oq"
	voiceId = "T5Cb98lTuWiTDQDY4AxZ"
	latency = 4
)

func TTS(text string) (*ffmpeg.AudioProvider, error) {

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s?optimize_streaming_latency=%v", voiceId, latency)

	//TODO: add previous_text
	//TODO: add previous_request_ids

	payloadString := fmt.Sprintf(`{
  "text": "%v",
  "voice_settings": {
    "stability": %v,
    "similarity_boost": %v,
    "use_speaker_boost": true
  },
  "model_id": "eleven_turbo_v2"
}`, text, stability, similarity)

	payload := strings.NewReader(payloadString)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("xi-api-key", os.Getenv("ELVENLABS_API_KEY"))

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("Response status for elevenlabs: " + res.Status)
	}

	opusProvider, err := ffmpeg.New(context.Background(), res.Body)

	if err != nil {
		return nil, err
	}
	return opusProvider, nil
}
