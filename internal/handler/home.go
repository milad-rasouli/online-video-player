package handler

import "github.com/gofiber/fiber/v2"

func GetHome(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
