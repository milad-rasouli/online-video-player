package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/websocket/v2"
)

type Message struct {
	Action string `json:"action"`
	Time   int    `json:"time,omitempty"`
	Src    string `json:"src,omitempty"`
}

var clients = make(map[*websocket.Conn]bool)

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
		// videoHandler := handler.Video{
		// 	Cfg: cfg,
		// }
		uploadHandler := handler.Upload{
			Cfg: cfg,
		}

		homeGroup := app.Group("/")
		authGroup := app.Group("/auth")
		// videoGroup := app.Group("/video")
		uploadGroup := app.Group("/upload")

		homeHandler.Register(homeGroup)
		authHandler.Register(authGroup)
		// videoHandler.Register(videoGroup)
		uploadHandler.Register(uploadGroup)
	}

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		clients[c] = true
		defer delete(clients, c)

		var msg Message
		for {
			if err := c.ReadJSON(&msg); err != nil {
				log.Println("read:", err)
				break
			}

			for client := range clients {
				if err := client.WriteJSON(msg); err != nil {
					log.Println("write:", err)
					break
				}
			}
		}
	}))

	app.Static("/static", "./internal/static")
	app.Listen(cfg.ProgramPort)
}
