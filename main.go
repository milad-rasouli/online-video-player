package main

import (
	"log"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
)

func main() {
	cfg := config.Config{}
	cfg.Read()
	log.Printf("%+v\n", cfg)
}
