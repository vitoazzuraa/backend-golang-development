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

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	StudentService *service.StudentService
	AuthService    *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	app.Get("/", func(c *fiber.Ctx) error {
		return helper.Ok(c, "API Students - PostgreSQL", nil)
	})

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)

		defer cancel()

		if err := deps.Pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}

		return helper.Ok(c, "server dan database berjalan", nil)
	})

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	studentRoutes := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	studentRoutes.Get("/", deps.StudentService.List)
	studentRoutes.Get("/:id", deps.StudentService.Get)
	studentRoutes.Post("/", deps.StudentService.Create)
	studentRoutes.Put("/:id", deps.StudentService.Replace)
	studentRoutes.Patch("/:id", deps.StudentService.Patch)
	studentRoutes.Delete("/:id", deps.StudentService.Delete)
}
