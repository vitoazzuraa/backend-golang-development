package route

import (
	"context"
	"time"

	"backend-go/api-student/app/service"
	"backend-go/api-student/helper"
	"backend-go/api-student/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Dependencies struct {
	Pool                *pgxpool.Pool
	Mongo               *mongo.Client
	JWT                 *helper.JWTManager
	Permissions         *helper.PermissionSet
	StudentService      *service.StudentService
	AuthService         *service.AuthService
	CourseService       *service.CourseService
	SupervisionService  *service.SupervisionService
	PortfolioService    *service.PortfolioService
	ReviewService       *service.ReviewService
	NotificationService *service.NotificationService
	AdminService        *service.AdminService
}

func Register(app *fiber.App, deps Dependencies) {
	app.Get("/", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, "API Students - PostgreSQL", nil)
	})

	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	studentRoutes := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	studentRoutes.Get("/",
		middleware.RequirePermission(perms, "student:list"),
		deps.StudentService.List)
	studentRoutes.Post("/",
		middleware.RequirePermission(perms, "student:create"),
		deps.StudentService.Create)
	studentRoutes.Delete("/:id",
		middleware.RequirePermission(perms, "student:delete"),
		deps.StudentService.Delete)

	studentRoutes.Get("/:id", deps.StudentService.Get)
	studentRoutes.Put("/:id", deps.StudentService.Replace)
	studentRoutes.Patch("/:id", deps.StudentService.Patch)

	studentRoutes.Get("/:id/portfolio", deps.PortfolioService.Get)
	studentRoutes.Put("/:id/portfolio", deps.PortfolioService.Save)
	studentRoutes.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
	studentRoutes.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.RemoveProject)

	portfolios := api.Group("/portfolios",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	portfolios.Get("/stats/technologies", deps.PortfolioService.TopTechnologies)

	reviews := api.Group("/reviews",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete)
	reviews.Post("/:id/replies", deps.ReviewService.Reply)
	reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)

	admin := api.Group("/admin",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
		middleware.RequirePermission(perms, "queue:monitor"))

	admin.Get("/queue", deps.AdminService.Queue)

	notifications := api.Group("/notifications",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	notifications.Get("/mine", deps.NotificationService.Mine)

	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"),
		deps.CourseService.UpdateQuota)

	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
	courses.Post("/:id/reviews", deps.ReviewService.Create)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	enrollments.Post("/swap", deps.CourseService.Swap)

	supervisors := api.Group("/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	supervisors.Get("/", deps.SupervisionService.List)
	supervisors.Post("/:id/assign",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisionService.Assign)
	supervisors.Patch("/:id",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisionService.UpdateQuota)

	supervisions := api.Group("/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	supervisions.Post("/switch",
		middleware.RequirePermission(perms, "supervision:manage"),
		deps.SupervisionService.Switch)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)

		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}

		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
