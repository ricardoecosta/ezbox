package main

type Metadata struct {
	Type string `json:"type"`
}

type ChannelChanged struct {
	Metadata
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Info    string `json:"info"`
}

func NewChannelChangedMessage(name string, picture string, info string) ChannelChanged {
	return ChannelChanged{
		Metadata: Metadata{Type: "ChannelChanged"},
		Name:     name,
		Picture:  picture,
		Info:     info}
}
