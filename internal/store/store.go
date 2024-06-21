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
	SaveUserVideoInfo(context.Context, model.User, model.VideoControllers) error
	GetUserVideoInfo(context.Context, model.User) (model.VideoControllers, error)
	SaveUploadedVideo(context.Context, model.UploadedVideo) error
	GetUploadedVideo(context.Context) (model.UploadedVideo, error)
	RemoveUploadedVideo(context.Context) error
	SaveDownloadVideoStatus(context.Context, model.DownloadStatus) error
	GetDownloadVideoStatus(context.Context) (model.DownloadStatus, error)
	RemoveDownloadVideoStatus(context.Context) error
}
