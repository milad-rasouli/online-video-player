package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	cfg := config.Config{}
	err := cfg.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", cfg)

	engine := html.New("internal/views/", ".html")

	if cfg.Debug == "true" {
		engine.Reload(true)
	} else {
		engine.Reload(false)
	}

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	{
		homeHandler := handler.Home{
			Cfg: cfg,
		}
		authHandler := handler.Auth{
			Cfg: cfg,
		}
		homeGroup := app.Group("/")
		authGroup := app.Group("/auth")

		homeHandler.Register(homeGroup)
		authHandler.Register(authGroup)
	}

	app.Static("/static", "./internal/static")
	app.Listen(cfg.ProgramPort)
}
