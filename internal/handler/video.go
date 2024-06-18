package handler

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/request"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/fiber/v2"
)

type Video struct {
	Cfg         config.Config
	redisClient store.RedisMessageStore
	movie       string
	timeline    string
	pause       bool
}

// whenever a client clicks on sync button it sends the status to this endpoint
func (u *Video) PostSetVideoControllers(c *fiber.Ctx) error {

	return nil
}
func (u *Video) PostGetVideoControllers(c *fiber.Ctx) error {
	var (
		vc  request.VideoControllers
		err error
	)
	err = c.BodyParser(&vc)
	if err != nil {
		log.Printf("GetVideoControllers json parse error %s", err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return nil
}

func (u *Video) PostUpload(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
func (u *Video) Register(c fiber.Router) {
	c.Post("/upload", u.PostUpload)
	c.Post("/set", u.PostSetVideoControllers)
	c.Post("/get", u.PostGetVideoControllers)
}
