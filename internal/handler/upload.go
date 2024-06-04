package handler

import (
	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Upload struct {
	Cfg config.Config
}

func (u *Upload) GetUpload(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
func (u *Upload) Register(c fiber.Router) {
	c.Get("/", u.GetUpload)
}
