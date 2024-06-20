package handler

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/Milad75Rasouli/online-video-player/internal/request"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
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

	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostSetVideoControllers invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	{
		log.Printf("set video %s", string(c.BodyRaw()))
		err = c.BodyParser(&vc)
		if err != nil {
			log.Printf("setVideoControllers body parse error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		err = vc.Valid()
		if err != nil {
			log.Printf("setVideoControllers body parse validation error %s", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if fullName != vc.User {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
	}

	{
		err = u.Store.SaveUserVideoInfo(c.Context(), model.User{FullName: vc.User}, model.VideoControllers{
			Pause:    vc.Pause,
			Timeline: vc.Timeline,
		})
		if err != nil {
			log.Printf("setVideoControllers save error %s", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	return nil
}
func (u *Video) PostGetVideoControllers(c *fiber.Ctx) error {
	var userFullName = c.Locals("userFullName")
	fullName, ok := userFullName.(string)
	if !ok {
		log.Printf("PostGetVideoControllers invalid userFullName error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	log.Printf("video PostGetVideo user is %s", fullName)
	usrVideoInto, err := u.Store.GetUserVideoInfo(c.Context(), model.User{FullName: fullName})
	if err != nil {
		log.Printf("PostGetVideoControllers store error")
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(usrVideoInto)
}

func (u *Video) PostUpload(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (u *Video) Video(c *fiber.Ctx) error {
	videoPath := "./internal/static/video/1.mkv"

	// Serve the video file with fasthttp.ServeFile
	fasthttp.ServeFile(c.Context(), videoPath)

	return nil
}

func (u *Video) Register(c fiber.Router) {
	c.Post("/upload", u.PostUpload)
	c.Post("/set", u.PostSetVideoControllers)
	c.Post("/get", u.PostGetVideoControllers)
	c.Get("/", u.Video)
}
