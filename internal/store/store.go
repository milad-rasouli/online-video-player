package store

import "github.com/Milad75Rasouli/online-video-player/internal/model"

type MessageStore interface {
	Save(model.Message) error
	GetAll() ([]model.Message, error)
}
