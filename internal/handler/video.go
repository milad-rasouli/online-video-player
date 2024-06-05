package handler

import (
	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Video struct {
	Cfg config.Config
}

func (u *Video) GetUpload(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
func (u *Video) Register(c fiber.Router) {
	c.Get("/", u.GetUpload)
}
