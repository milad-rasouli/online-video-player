package handler

import (
	"log"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/jwt"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	Cfg config.Config
	JWT jwt.AccessJWT
}

func (a *Auth) GetEntrance(c *fiber.Ctx) error {
	{
		accessToken := c.Cookies("_token")
		if len(accessToken) != 0 {
			return a.redirectToHome(c)
		}
	}
	return c.Render("entrance", fiber.Map{
		"Title": "Entrance",
	})
}

func (a *Auth) PostEntrance(c *fiber.Ctx) error {
	var (
		fullName = c.FormValue("name")
		password = c.FormValue("password")
		token    string
		err      error
	)
	log.Println("From /auth got this ", fullName, " ", password, " ", a.Cfg.Password, len(password), " ", len(a.Cfg.Password))
	{
		if len(fullName) == 0 || len(password) == 0 {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if a.Cfg.Password != password {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
	}
	{
		token, err = a.JWT.Create(model.User{FullName: fullName})
		if err != nil {
			log.Println("create JWT error", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		a.setTokenCookie(c, token)
	}
	log.Println("token ", token)
	return a.redirectToHome(c)
}

func (a *Auth) redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/home", fiber.StatusSeeOther)
}

func (a *Auth) POSTUpdateToken(c *fiber.Ctx) error {

	accessToken := c.Cookies("_token")
	if len(accessToken) == 0 {
		return a.removeCookiesAndRedirectToEntrance(c)
	} else {
		user, err := a.JWT.VerifyParse(accessToken)
		if err != nil {
			log.Println("failed to parse access token", err)
			return a.removeCookiesAndRedirectToEntrance(c)
		}
		token, err := a.JWT.Create(user)
		if err != nil {
			log.Println("failed to create access token", err)
			return a.removeCookiesAndRedirectToEntrance(c)
		}
		a.setTokenCookie(c, token)
		return c.SendStatus(fiber.StatusCreated)
	}
}

func (a *Auth) setTokenCookie(c *fiber.Ctx, token string) {
	var (
		expTime time.Time
		path    string
		name    string
	)

	expTime = time.Now().Add(time.Minute * time.Duration(a.Cfg.JWtExpireTime))
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

func (a *Auth) UserMiddleWare(c *fiber.Ctx) error {
	var (
		accessToken string
		err         error
		user        model.User
	)
	{
		accessToken = c.Cookies("_token")
		notUser := len(accessToken) == 0
		if notUser == true {
			return a.removeCookiesAndRedirectToEntrance(c)
		}
	}
	{
		user, err = a.JWT.VerifyParse(accessToken)
		if err != nil {
			log.Println("failed to parse access token", err)
			return a.removeCookiesAndRedirectToEntrance(c)
		}
	}

	log.Println("User Middleware", user.FullName)
	c.Locals("userFullName", user.FullName)
	return c.Next()
}

func (a *Auth) removeCookiesAndRedirectToEntrance(c *fiber.Ctx) error {
	c.ClearCookie("_token")
	return c.Redirect("/auth", fiber.StatusSeeOther)
}

func (a *Auth) Register(c fiber.Router) {
	c.Get("/", a.GetEntrance)
	c.Post("/", a.PostEntrance)
	c.Post("/update", a.POSTUpdateToken)
}
