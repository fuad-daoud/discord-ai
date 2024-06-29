package db

type Guild struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type VoiceChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TextChannel struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CreatedDate string `json:"createdDate"`
}

type Thread struct {
	TextChannel
}

type Message struct {
	Id          string `json:"id"`
	Text        string `json:"text"`
	CreatedDate string `json:"createdDate"`
	UpdatedDate string `json:"updatedDate"`
}

type Member struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatarUrl"`
}
