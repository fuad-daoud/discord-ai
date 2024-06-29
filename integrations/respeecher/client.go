package respeecher

import (
	"fmt"
	"github.com/forPelevin/gomoji"
	"github.com/fuad-daoud/discord-ai/integrations/custom_http"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ParentFolderId = "dbb22f4f-6e0e-46ac-a161-eaa57f094e32"
	Label          = "bot_txt"
	Alice          = Voice{
		Id: "78e29195-7ab3-46ab-9d72-1ecfee9ab076",
		Accents: map[string]string{
			"NoAccent":        "cbd490ac-86ab-4618-b2e8-6d104a8402f5",
			"EnBalanceAccent": "d2fd41ae-fd00-46f3-a89d-baef6223e8ce",
			"EnUSAccent":      "dcc893a1-dfdf-4bac-97db-0dc90a335d50",
		},
		Styles: map[string]string{
			"NeutralYouthfulNarration": "84f40abd-e8e4-47de-9ac5-9cf277e0d4a2",
			"LivelyAnimatedNarration":  "8dbae598-b777-4207-8634-c42ade4f977b",
			"AstonishedNarration":      "0cd420c0-6786-4a82-ba07-621021204939",
		},
	}
	Oksana = Voice{
		Id: "1a1beed0-d80e-4b1c-b93d-7a8a50232f01",
		Accents: map[string]string{
			"NoAccent":        "a8c44794-0a88-4442-9c79-e1df4e2c8368",
			"EnBalanceAccent": "b40c8fcb-c29a-4c83-9bcb-1f5f5e57b339",
			"EnUSAccent":      "a9c43dfa-30e3-47aa-bf5d-20ee1a71f35f",
		},
		Styles: map[string]string{
			"IntenseFirm":         "2308aa2c-14ed-44bb-91de-8ceaf73eaf03",
			"HushedRaspy":         "a73fc00c-623a-436c-af29-eb2c3b408da3",
			"Amazed":              "5bc22ee7-d08a-43b5-a73c-0972e350f93f",
			"IrritatedDetermined": "c8e1ac15-f307-471b-ad64-32c4e4f3d1f2",
			"StrainedWhispering":  "a153f1e9-9557-4682-acc4-11141c2a7726",
			"EnergeticGiggly":     "b11c1a58-c67a-4df9-853b-de21b9857670",
			"AstonishedUpbeat":    "21a4be0b-af5e-4993-b98f-da3c1b43ad1f",
			"BubblyLively":        "68f5c405-c321-42d4-92a0-1b45ea83bde4",
		},
	}
	OksanaDefault = VoiceParams{
		Id:     Oksana.Id,
		Accent: Oksana.Accents["EnUSAccent"],
		Style:  Oksana.Styles["HushedRaspy"],
	}
	Laura = Voice{
		Id: "f812c129-0497-4126-a558-037a372052cf",
		Accents: map[string]string{
			"NoAccent":        "1489de3e-a82e-4ec4-9c2c-a3836f887f85",
			"EnBalanceAccent": "cf308708-6f05-41d3-85bc-1cc6b48a5286",
			"EnUSAccent":      "b784b369-8ccc-4f08-afd8-9c8d07a70c11",
		},
		Styles: map[string]string{
			"EnticingBreathyRender": "72b59672-720c-4145-a05e-644e26dcdaaa",
			"EnergeticStormy":       "dbe0a62d-cc4a-4fa2-a73f-71ef02b72e54",
			"SassyBoastful":         "d32b61a4-60fc-4b68-b3ad-2ed543d372a8",
			"PositiveRelaxed":       "2517b12a-0bea-4d7c-9531-108c46fa853a",
			"AnimatedBold":          "2dd911e4-ba8b-4fad-9fb8-cb66d6afede8",
			"LivelyAmazed":          "5eecf882-03f8-48ae-8da9-475152e24820",
			"Upbeat":                "e2044f2e-8dda-4fac-a2b6-a97d48df4412",
		},
	}
	LauraDefault = VoiceParams{
		Id:     Laura.Id,
		Accent: Laura.Accents["EnUSAccent"],
		Style:  Laura.Styles["PositiveRelaxed"],
	}
	Text     = "The song Shape of You is playing now. Enjoy every beat, darling."
	FilePath = "test.wav"
)

type Voice struct {
	Id      string
	Accents map[string]string
	Styles  map[string]string
}
type VoiceParams struct {
	Id     string
	Accent string
	Style  string
}

type Client interface {
	DefaultTextToSpeech(text string) (string, error)
	TextToSpeech(text string, voice VoiceParams) (string, error)
	createOrder(result Response, voice VoiceParams) []Response
	createTTS(text string) Response
	getRecording(secondResult []Response) Response
	getRecordings(secondResult []Response, recording *Response)
}

type defaultClient struct {
	Client custom_http.Client
}

var client Client

func GetClient() Client {
	if client == nil {
		client = makeClient()
		return client
	}
	return client
}

func makeClient() Client {
	headers := make(map[string]string)
	headers["accept"] = "application/json"
	headers["api-key"] = os.Getenv("RESPEECHER_API_KEY")
	headers["Content-Type"] = "application/json"
	var httpClient custom_http.Client = &custom_http.DefaultClient{
		BaseURL: "https://gateway.respeecher.com",
		Client:  &http.Client{},
		Headers: headers,
	}

	client = &defaultClient{
		Client: httpClient,
	}
	return client
}

func (dc *defaultClient) DefaultTextToSpeech(text string) (string, error) {
	return dc.TextToSpeech(text, OksanaDefault)
}

func (dc *defaultClient) TextToSpeech(text string, voiceParams VoiceParams) (string, error) {

	result := dc.createTTS(text)

	secondResult := dc.createOrder(result, voiceParams)

	recording := dc.getRecording(secondResult)

	req := dc.Client.GetRequest(recording.Url)

	bytes := dc.Client.Do(req)
	filePath := getFilePath()
	err := os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func getFilePath() string {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	file := "files/wav/" + newUUID.String() + ".wav"
	return file
}

func (dc *defaultClient) createOrder(result Response, voice VoiceParams) []Response {
	slog.Info("voice params ", "voiceParams", voice)
	data := strings.NewReader(fmt.Sprintf(`{
  "original_id": "%s",
  "conversions": [
    {
      "voice_id": "%s",
      "narration_style_id": "%s",
      "accent_id": "%s",
      "semitones_correction": 0
    }
  ]
}`, result.Id, voice.Id, voice.Style, voice.Accent))

	var secondResult []Response
	req := dc.Client.PostRequest("/api/v2/orders", data)
	dc.Client.DoJson(req, &secondResult)
	return secondResult
}

func (dc *defaultClient) createTTS(text string) Response {
	text = gomoji.RemoveEmojis(text)
	var data = strings.NewReader(fmt.Sprintf(`{
  "parent_folder_id": "%s",
  "text": "... %s ...",
  "label": "%s"
}`, ParentFolderId, text, Label))
	req := dc.Client.PostRequest("/api/v2/recordings/tts", data)
	var result Response
	dc.Client.DoJson(req, &result)
	return result
}

func (dc *defaultClient) getRecording(secondResult []Response) Response {
	var recording Response
	var counter = 0
	dc.getRecordings(secondResult, &recording)
	counter++
	for {
		if recording.State == "done" {
			break
		}
		if counter == 60 {
			slog.Error("could not find ready recording")
		}
		time.Sleep(200 * time.Millisecond)
		dc.getRecordings(secondResult, &recording)
		counter++
	}
	return recording
}

func (dc *defaultClient) getRecordings(secondResult []Response, recording *Response) {
	req := dc.Client.GetRequest("/api/recordings?folder_id=dbb22f4f-6e0e-46ac-a161-eaa57f094e32&limit=50&offset=0&direction=desc")

	var recordings Recordings
	dc.Client.DoJson(req, &recordings)

	slog.Info("Finding the record ", "ID", secondResult[0].OriginalId)

	for i := 0; i < len(recordings.List); i++ {
		if recordings.List[i].OriginalId == secondResult[0].OriginalId {
			recording.Id = recordings.List[i].Id
			recording.Url = recordings.List[i].Url
			recording.OriginalId = recordings.List[i].OriginalId
			recording.State = recordings.List[i].State
			break
		}
	}
}

type Response struct {
	Id         string `json:"id"`
	OriginalId string `json:"original_id"`
	State      string `json:"state"`
	Url        string `json:"url"`
}

type Recordings struct {
	List []Response `json:"list"`

	Pagination struct {
		Count  int `json:"count"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
}
