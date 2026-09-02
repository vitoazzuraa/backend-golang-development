package main

import (
	"context"
	"log"
	"mime"
	"time"

	"backend-go/api-student/config"
	"backend-go/api-student/database"
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
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

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
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return ok(c, "server dan database berjalan", nil)
	})

	studentRoutes := api.Group("/students", requireJSON)
	studentRoutes.Get("/", listStudents)
	studentRoutes.Get("/:id", getStudent)
	studentRoutes.Post("/", createStudent)
	studentRoutes.Put("/:id", replaceStudent)
	studentRoutes.Patch("/:id", patchStudent)
	studentRoutes.Delete("/:id", deleteStudent)

	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}
