package request

import (
	"errors"
)

var (
	TimelineIsEmpty     = errors.New("Timeline can't be empty")
	UserIsEmpty         = errors.New("Movie can't be empty")
	UploadedURLIsEmpty  = errors.New("Uploaded URL can't be empty")
	UploadedUUIDIsEmpty = errors.New("Uploaded URL can't be empty")
)

type VideoControllers struct {
	Pause    bool   `json:"pause,omitempty"`
	Timeline string `json:"timeline,omitempty"`
	User     string `json:"user"`
}

func (v *VideoControllers) Valid() error {
	if len(v.Timeline) == 0 {
		return TimelineIsEmpty
	}
	if len(v.User) == 0 {
		return UserIsEmpty
	}
	return nil
}

type UploadedVideo struct {
	URL  string `json:"url"`
	UUID string `json:"uuid"`
}

func (u *UploadedVideo) Valid() error {
	if len(u.URL) == 0 {
		return UploadedURLIsEmpty
	}
	if len(u.UUID) == 0 {
		return UploadedUUIDIsEmpty
	}
	return nil
}
