package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Password      string
	ProgramPort   string
	Debug         string
	WebsiteTitle  string
	RedisAddress  string
	RedisChatExp  string
	UserPassword  string
	JWTSecret     string
	JWtExpireTime uint64
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
	c.UserPassword = os.Getenv("USER_PASSWORD")
	c.JWTSecret = os.Getenv("JWT_SECRET")
	c.JWtExpireTime, err = strconv.ParseUint(os.Getenv("JWT_EXPIRE_TIME"), 10, 64)
	if err != nil {
		return err
	}

	if len(c.RedisChatExp) == 0 {
		c.RedisChatExp = "60"
	}
	if len(c.UserPassword) == 0 {
		c.UserPassword = "1234qwer"
	}
	if len(c.JWTSecret) == 0 {
		c.JWTSecret = "32t8ndoiufs9bgSGerwon#fdsg"
	}
	if c.JWtExpireTime <= 0 {
		c.JWtExpireTime = 1
	}

	return nil
}
