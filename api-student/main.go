package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"backend-go/api-student/app/repository"
	"backend-go/api-student/app/service"
	"backend-go/api-student/config"
	"backend-go/api-student/database"
	"backend-go/api-student/helper"
	"backend-go/api-student/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()

	logger := config.NewLogger()

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

	txManager := repository.NewTxManager(pool)

	courseService := service.NewCourseService(courseRepository, txManager)

	supervisionService := service.NewSupervisionService(supervisorRepository, txManager)

	portfolioRepository := repository.NewPortfolioRepository(mongoDB)

	if err := portfolioRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal menyiapkan index portofolio", slog.String("error", err.Error()))
		os.Exit(1)
	}

	portfolioService := service.NewPortfolioService(studentRepository, portfolioRepository)

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
		Pool:               pool,
		Mongo:              mongoClient,
		JWT:                jwtManager,
		Permissions:        permissions,
		StudentService:     studentService,
		AuthService:        authService,
		CourseService:      courseService,
		SupervisionService: supervisionService,
		PortfolioService:   portfolioService,
	})

	logger.Info("server berjalan")

	port := config.GetEnv("APP_PORT", "3000")

	log.Fatal(app.Listen(":" + port))
}
