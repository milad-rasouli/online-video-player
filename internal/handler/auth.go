package handler

import (
	"log"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	Cfg config.Config
}

func (a *Auth) GetEntrance(c *fiber.Ctx) error {
	return c.Render("entrance", fiber.Map{
		"Title": "Entrance",
	})
}

func (a *Auth) PostEntrance(c *fiber.Ctx) error {
	var (
		name     = c.FormValue("name")
		password = c.FormValue("password")
	)
	if len(name) == 0 || len(password) == 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if a.Cfg.Password == password {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	log.Println("From /auth got this ", name, " ", password)
	return c.Redirect("/", fiber.StatusSeeOther)
}

func (a *Auth) POSTUpdateToken(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func (a *Auth) SetTokenCookie(c *fiber.Ctx, token string) {
	var (
		expTime time.Time
		path    string
		name    string
	)

	expTime = time.Now().Add(time.Second * time.Duration(a.Cfg.JWtExpireTime))
	path = "/"
	name = "_token"
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    token,
		Expires:  expTime,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     path,
		// Domain:   "online-video-player.ir", //TODO: take it from the configuration
	})
}

func (a *Auth) Register(c fiber.Router) {
	c.Get("/", a.GetEntrance)
	c.Post("/", a.PostEntrance)
	c.Post("/update-token", a.POSTUpdateToken)
}
