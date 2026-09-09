package middleware

import (
	"log/slog"
	"mime"
	"time"

	"backend-go/api-student/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())
	app.Use(RequestLogger(logger))
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)

		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

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
