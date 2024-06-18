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
	Save(context.Context, model.VideoControllers) error
	Get(context.Context) (model.VideoControllers, error)
}
