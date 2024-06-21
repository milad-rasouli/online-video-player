package model

import "time"

type VideoControllers struct {
	Pause    bool   `json:"pause"`
	Timeline string `json:"timeline"`
}

type UploadedVideo struct {
	URL  string
	UUID string
}

type DownloadStatus struct {
	TotalSize    uint64
	ReceivedSize uint64
	StartTime    time.Time
}
