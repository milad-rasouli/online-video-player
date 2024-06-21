package model

type VideoControllers struct {
	Pause    bool   `json:"pause"`
	Timeline string `json:"timeline"`
}

type UploadedVideo struct {
	URL  string
	UUID string
}

type DownloadStatus struct {
	User         string  `json:"user"`
	TotalSize    uint64  `json:"totalSize"`
	ReceivedSize uint64  `json:"receivedSize"`
	StartTime    int64   `json:"startTime"`
	Percent      float64 `json:"percent"`
	Speed        float64 `json:"speed"`
	TimeLeft     string  `json:"timeLeft"`
}
