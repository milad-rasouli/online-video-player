package store

import (
	"context"

	"github.com/Milad75Rasouli/online-video-player/internal/model"
)

type MessageStore interface {
	Save(context.Context, model.Message) error
	GetAll(context.Context) ([]model.Message, error)
}

type UserAndVideoStore interface {
	SaveUser(context.Context, model.User) error
	RemoveAllUser(context.Context) error
	GetAllUser(context.Context) ([]model.User, error)
	SaveCurrentVideo(context.Context, model.VideoControllers) error
	GetCurrentVideo(context.Context, model.User) (model.VideoControllers, error)
	RemoveCurrentVideo(context.Context, model.User) error
	SaveToPlaylist(context.Context, model.Playlist) error
	GetPlaylist(context.Context) ([]model.Playlist, error)
	RemovePlaylist(context.Context) error
}
