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
	RedisChatExp string
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
	c.RedisChatExp = os.Getenv("REDIS_CHAT_EXP")

	if len(c.RedisChatExp) == 0 {
		c.RedisChatExp = "60"
	}

	return nil
}
