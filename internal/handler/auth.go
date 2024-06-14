package handler

import (
	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	Cfg config.Config
}

func (a *Auth) GetSignIn(c *fiber.Ctx) error {
	return c.Render("sign-in", fiber.Map{
		"Title": a.Cfg.WebsiteTitle + " Sign-in", //TODO: take from .env
	})
}

func (a Auth) POSTUpdateToken(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (a *Auth) Register(c fiber.Router) {
	c.Get("/", a.GetSignIn)
	c.Post("/update-token", a.POSTUpdateToken)
}
