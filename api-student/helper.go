package main

import "github.com/gofiber/fiber/v2"

func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}
