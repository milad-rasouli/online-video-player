package model

type VideoControllers struct {
	Pause    bool   `json:"pause"`
	Timeline string `json:"timeline"`
	Movie    string `json:"movie"`
}

type Playlist struct {
	Item string
}
