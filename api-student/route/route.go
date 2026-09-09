package route

import (
	"context"
	"time"
	
	"backend-go/api-student/app/service"
	"backend-go/api-student/helper"
	"backend-go/api-student/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	app.Get("/", func(c *fiber.Ctx) error {
		return helper.Ok(c, "API Students - PostgreSQL", nil)
	})

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)

		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return helper.Ok(c, "server dan database berjalan", nil)
	})

	studentRoutes := api.Group("/students", middleware.RequireJSON)
	studentRoutes.Get("/", studentService.List)
	studentRoutes.Get("/:id", studentService.Get)
	studentRoutes.Post("/", studentService.Create)
	studentRoutes.Put("/:id", studentService.Replace)
	studentRoutes.Patch("/:id", studentService.Patch)
	studentRoutes.Delete("/:id", studentService.Delete)
}