package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Password     string
	ProgramPort  string
	Debug        string
	WebsiteTitle string
	RedisAddress string
}

func (c *Config) Read() error {
	var (
		err error
	)
	err = godotenv.Load()
	if err != nil {
		return err
	}
	c.Password = os.Getenv("USER_PASSWORD")
	c.ProgramPort = os.Getenv("PROGRAM_PORT")
	c.Debug = os.Getenv("DEBUG")
	c.WebsiteTitle = os.Getenv("WEBSITE_TITLE")
	c.RedisAddress = os.Getenv("REDIS_ADDRESS")
	return nil
}
