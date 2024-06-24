package database

type Guild struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type VoiceChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TextChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}
