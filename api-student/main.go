package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

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

	app.Get("/", func(c *fiber.Ctx) error {
		return ok(c, "API Students - Task 3", nil)
	})

	api := app.Group("/api/v1")
	studentRoutes := api.Group("/students")
	studentRoutes.Get("/", listStudents)
	studentRoutes.Get("/:id", getStudent)
	studentRoutes.Post("/", createStudent)
	studentRoutes.Put("/:id", replaceStudent)
	studentRoutes.Patch("/:id", patchStudent)
	studentRoutes.Delete("/:id", deleteStudent)

	log.Fatal(app.Listen(":3000"))
}
