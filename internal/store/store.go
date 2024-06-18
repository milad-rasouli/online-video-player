package store

import (
	"context"

	"github.com/Milad75Rasouli/online-video-player/internal/model"
)

type MessageStore interface {
	Save(context.Context, model.Message) error
	GetAll(context.Context) ([]model.Message, error)
}

type VideoControllersStore interface {
	SaveCurrentVideo(context.Context, model.VideoControllers) error
	GetCurrentVideo(context.Context) (model.VideoControllers, error)
	SaveToPlaylist(context.Context, model.Playlist) error
	GetPlaylist(context.Context) ([]model.Playlist, error)
	RemovePlaylist(context.Context) error
}
