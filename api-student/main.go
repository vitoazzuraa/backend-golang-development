package main

import (
	"context"
	"log"

	"backend-go/api-student/app/repository"
	"backend-go/api-student/app/service"
	"backend-go/api-student/config"
	"backend-go/api-student/database"
)

func main() {
	config.LoadEnv()

	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())

	if err != nil {
		logger.Error("gagal terhubung ke database")
		log.Fatal(err)
	}

	defer pool.Close()

	studentRepository := repository.NewStudentRepository(pool)

	studentService := service.NewStudentService(studentRepository)

	app := config.NewApp(logger, pool, studentService)

	port := config.GetEnv("APP_PORT", "3000")

	log.Fatal(app.Listen(":" + port))
}
