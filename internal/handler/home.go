package handler

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Home struct {
	Cfg config.Config
}

func (h *Home) GetHome(c *fiber.Ctx) error {
	var (
		userFullName = c.Locals("userFullName")
	)
	log.Println("home page user full name ", userFullName)
	return c.Render("home", fiber.Map{
		"Title":   h.Cfg.WebsiteTitle,
		"Name":    userFullName,
		"JWTTime": h.Cfg.JWtExpireTime / 6,
		"Debug":   h.Cfg.Debug,
	})
}
func (h *Home) Register(c fiber.Router) {
	c.Get("/", h.GetHome)
}
