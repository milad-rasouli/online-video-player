package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Config{}
	err := cfg.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", cfg)

	app := fiber.New(fiber.Config{})
	app.Get("/", handler.GetHome)

	app.Listen(cfg.ProgramPort)
}
