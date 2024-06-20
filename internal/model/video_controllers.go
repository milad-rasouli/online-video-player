package model

type VideoControllers struct {
	Pause    bool   `json:"pause"`
	Timeline string `json:"timeline"`
}

type UploadedVideo struct {
	URL  string
	UUID string
}
