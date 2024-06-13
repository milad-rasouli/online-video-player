package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/handler"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
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

	messageStore, err := store.NewRedisMessageStore(cfg)

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
		videoHandler := handler.Video{
			Cfg: cfg,
		}
		uploadHandler := handler.Upload{
			Cfg: cfg,
		}
		chatHandler := handler.NewChat(cfg, messageStore)

		homeGroup := app.Group("/")
		authGroup := app.Group("/auth")
		videoGroup := app.Group("/video")
		uploadGroup := app.Group("/upload")
		ChatGroup := app.Group("/chat")

		homeHandler.Register(homeGroup)
		authHandler.Register(authGroup)
		videoHandler.Register(videoGroup)
		uploadHandler.Register(uploadGroup)
		chatHandler.Register(ChatGroup)
	}

	app.Static("/static", "./internal/static")
	log.Fatal(app.Listen(cfg.ProgramPort))
}
