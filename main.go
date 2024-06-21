package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/handler"
	"github.com/Milad75Rasouli/online-video-player/internal/jwt"
	"github.com/Milad75Rasouli/online-video-player/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "go.uber.org/automaxprocs"
)

func main() {

	cfg := config.Config{}
	err := cfg.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", cfg)

	messageStore, disposeRedis, err := store.NewRedisMessageStore(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer disposeRedis()
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
			JWT: jwt.NewAccessJWT(cfg),
		}
		userAndVideo, cancel, err := store.NewRedisUserAndVideStore(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer cancel()
		videoHandler := handler.Video{
			Cfg:   cfg,
			Store: userAndVideo,
		}
		chatHandler := handler.NewChat(cfg, messageStore)

		app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/home", fiber.StatusTemporaryRedirect) })
		homeGroup := app.Group("/home", authHandler.UserMiddleWare)
		authGroup := app.Group("/auth")
		videoGroup := app.Group("/video", authHandler.UserMiddleWare)
		ChatGroup := app.Group("/chat", authHandler.UserMiddleWare)

		homeHandler.Register(homeGroup)
		authHandler.Register(authGroup)
		videoHandler.Register(videoGroup)
		chatHandler.Register(ChatGroup)
	}

	app.Static("/static", "./internal/static")
	log.Fatal(app.Listen(cfg.ProgramPort))
}
