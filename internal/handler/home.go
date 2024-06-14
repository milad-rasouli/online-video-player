package handler

import (
	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Home struct {
	Cfg config.Config
}

func (h *Home) GetHome(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{
		"Title": h.Cfg.WebsiteTitle,
		"Name":  "Milad",
	})
}
func (h *Home) Register(c fiber.Router) {
	c.Get("/", h.GetHome)
}
