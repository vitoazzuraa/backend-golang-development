package middleware

import (
	"mime"

	"backend-go/api-student/helper"
	"github.com/gofiber/fiber/v2"
)

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if !methodsWithBody[c.Method()] {
		return c.Next()
	}

	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))

	if err != nil || mediaType != fiber.MIMEApplicationJSON {
		return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}

	return c.Next()
}