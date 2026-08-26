package main

import (
	"log"
	"mime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func requireJSON(c *fiber.Ctx) error {
	if !metodeBerbody[c.Method()] {
		return c.Next()
	}

	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || mediaType != fiber.MIMEApplicationJSON {
		return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}

	return c.Next()
}

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			}
			return fail(c, status, err.Error())
		},
	})
	app.Use(requestid.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return ok(c, "API Students - Task 4", nil)
	})

	api := app.Group("/api/v1")
	studentRoutes := api.Group("/students", requireJSON)
	studentRoutes.Get("/", listStudents)
	studentRoutes.Get("/:id", getStudent)
	studentRoutes.Post("/", createStudent)
	studentRoutes.Put("/:id", replaceStudent)
	studentRoutes.Patch("/:id", patchStudent)
	studentRoutes.Delete("/:id", deleteStudent)

	log.Fatal(app.Listen(":3000"))
}
