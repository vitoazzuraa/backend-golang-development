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

	studentRepository := repository.NewStudentRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)

	studentService := service.NewStudentService(studentRepository)

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	app := config.NewApp(logger, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		StudentService: studentService,
		AuthService:    authService,
	})

	logger.Info("server berjalan")

	port := config.GetEnv("APP_PORT", "3000")

	log.Fatal(app.Listen(":" + port))
}
