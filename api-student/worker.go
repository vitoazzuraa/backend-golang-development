package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/app/service"
	"backend-go/api-student/config"
	"backend-go/api-student/helper"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const mailLogPath = "logs/mail.log"

func startWorker(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) func() {
	mailer, err := helper.NewLogMailer(mailLogPath,
		time.Duration(config.GetEnvInt("MAILER_DELAY_MS", 1500))*time.Millisecond,
		config.GetEnvInt("MAILER_FAIL_EVERY", 0))

	if err != nil {
		logger.Error("gagal menyiapkan mailer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	outboxRepository := repository.NewOutboxRepository(pool)
	notificationRepository := repository.NewNotificationRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	txManager := repository.NewTxManager(pool)

	client := asynq.NewClient(config.RedisOpt())

	relay := service.NewOutboxRelay(outboxRepository, notificationRepository, txManager, client, logger,
		time.Duration(config.GetEnvInt("RELAY_INTERVAL_MS", 1000))*time.Millisecond,
		config.GetEnvInt("RELAY_BATCH", 100))

	worker := service.NewNotificationWorker(userRepository, notificationRepository, txManager, mailer, logger)

	mux := asynq.NewServeMux()
	mux.HandleFunc(model.TaskSendNotification, worker.HandleSend)

	srv := config.NewWorkerServer(logger)

	if err := srv.Start(mux); err != nil {
		logger.Error("worker gagal dimulai", slog.String("error", err.Error()))
		os.Exit(1)
	}

	relayCtx, stopRelay := context.WithCancel(ctx)

	relayDone := make(chan struct{})

	go func() {
		defer close(relayDone)

		relay.Run(relayCtx)
	}()

	var once sync.Once

	return func() {
		once.Do(func() {
			stopRelay()

			<-relayDone

			srv.Shutdown()

			_ = client.Close()
		})
	}
}
