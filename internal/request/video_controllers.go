package request

import (
	"errors"
)

var (
	TimelineIsEmpty = errors.New("Timeline can't be empty")
	MovieIsEmpty    = errors.New("Movie can't be empty")
)

type VideoControllers struct {
	Pause    bool   `json:"pause"`
	Timeline string `json:"timeline"`
	Movie    string `json:"movie"`
}

func (v *VideoControllers) Valid() error {
	if len(v.Timeline) == 0 {
		return TimelineIsEmpty
	}
	if len(v.Movie) == 0 {
		return MovieIsEmpty
	}
	return nil
}
