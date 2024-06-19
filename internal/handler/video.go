package handler

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/Milad75Rasouli/online-video-player/internal/request"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Video struct {
	Cfg   config.Config
	Store store.UserAndVideoStore
}

func (u *Video) PostSetVideoControllers(c *fiber.Ctx) error {
	var (
		vc  request.VideoControllers
		err error
	)
	{
		err = c.BodyParser(&vc)
		if err != nil {
			log.Printf("setVideoControllers body parse error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		log.Printf("set vc %+v\n", vc) //TODO: Delete this line

		err = vc.Valid()
		if err != nil {
			log.Printf("setVideoControllers body parse validation error %s", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	{
		err = u.Store.SaveCurrentVideo(c.Context(), model.VideoControllers{
			Pause:    vc.Pause,
			Timeline: vc.Timeline,
			Movie:    vc.Movie,
		})
		if err != nil {
			log.Printf("setVideoControllers save error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	return nil
}
func (u *Video) PostGetVideoControllers(c *fiber.Ctx) error {
	// var (
	// 	vc  request.VideoControllers
	// 	err error
	// )
	// err =
	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostGetVideoControllers invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	log.Printf("video PostGetVideo user is %s", fullName)
	list, err := u.Store.GetCurrentVideo(c.Context(), model.User{FullName: fullName})
	if err != nil {

		log.Printf("PostGetVideoControllers store error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(list)
}

func (u *Video) PostUpload(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
func (u *Video) Register(c fiber.Router) {
	c.Post("/upload", u.PostUpload)
	c.Post("/set", u.PostSetVideoControllers)
	c.Post("/get", u.PostGetVideoControllers)
}
