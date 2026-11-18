package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-go/api-student/app/repository"
	"backend-go/api-student/app/service"
	"backend-go/api-student/config"
	"backend-go/api-student/database"
	"backend-go/api-student/helper"
	"backend-go/api-student/route"

	"github.com/hibiken/asynq"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()

	logger := config.NewLogger()

	role := flag.String("role", config.GetEnv("APP_ROLE", "all"), "peran proses: api, worker, atau all")

	flag.Parse()

	jwtSecret := config.GetEnv("JWT_SECRET", "")

	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())

	if err != nil {
		logger.Error("gagal terhubung ke database")
		log.Fatal(err)
	}

	defer pool.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	stopWorker := func() {}

	if *role != "api" {
		stopWorker = startWorker(ctx, logger, pool)
	}

	if *role == "worker" {
		<-ctx.Done()

		stopWorker()

		return
	}

	mongoClient, err := database.NewMongoClient(context.Background())

	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		_ = mongoClient.Disconnect(ctx)
	}()

	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "go_backend"))

	studentRepository := repository.NewStudentRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	courseRepository := repository.NewCourseRepository(pool)

	supervisorRepository := repository.NewSupervisorRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())

	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)

	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentService := service.NewStudentService(studentRepository, permissions)

	outboxRepository := repository.NewOutboxRepository(pool)

	notificationRepository := repository.NewNotificationRepository(pool)

	txManager := repository.NewTxManager(pool)

	courseService := service.NewCourseService(courseRepository, outboxRepository, txManager)

	supervisionService := service.NewSupervisionService(supervisorRepository, txManager)

	portfolioRepository := repository.NewPortfolioRepository(mongoDB)

	if err := portfolioRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal menyiapkan index portofolio", slog.String("error", err.Error()))
		os.Exit(1)
	}

	portfolioService := service.NewPortfolioService(studentRepository, portfolioRepository)

	reviewRepository := repository.NewReviewRepository(mongoDB)

	if err := reviewRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal menyiapkan index ulasan", slog.String("error", err.Error()))
		os.Exit(1)
	}

	reviewService := service.NewReviewService(reviewRepository, courseRepository, permissions)

	notificationService := service.NewNotificationService(notificationRepository)

	adminService := service.NewAdminService(outboxRepository, asynq.NewInspector(config.RedisOpt()))

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	app := config.NewApp(logger, route.Dependencies{
		Pool:                pool,
		Mongo:               mongoClient,
		JWT:                 jwtManager,
		Permissions:         permissions,
		StudentService:      studentService,
		AuthService:         authService,
		CourseService:       courseService,
		SupervisionService:  supervisionService,
		PortfolioService:    portfolioService,
		ReviewService:       reviewService,
		NotificationService: notificationService,
		AdminService:        adminService,

		EnableSwagger: os.Getenv("SWAGGER_ENABLED") == "true",
	})

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
		}
	}()

	logger.Info("server berjalan")

	<-ctx.Done()

	if err := app.Shutdown(); err != nil {
		logger.Error("gagal menghentikan server", slog.String("error", err.Error()))
	}

	stopWorker()
}
