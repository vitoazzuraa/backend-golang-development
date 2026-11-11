package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"backend-go/api-student/app/model"

	"github.com/hibiken/asynq"
)

const maxTaskAttempts = 6

func RedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:         GetEnv("REDIS_ADDR", "localhost:6379"),
		Password:     GetEnv("REDIS_PASSWORD", ""),
		DB:           GetEnvInt("REDIS_DB", 0),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

func NewWorkerServer(logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(RedisOpt(), asynq.Config{
		Concurrency: GetEnvInt("WORKER_CONCURRENCY", 5),
		Queues:      map[string]int{model.QueueNotifications: 1},
		RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
			return time.Duration(n+1) * 2 * time.Second
		},
		ShutdownTimeout: 10 * time.Second,
		ErrorHandler:    newTaskErrorHandler(logger),
		Logger:          newAsynqLogger(logger),
		LogLevel:        asynq.WarnLevel,
	})
}

func newTaskErrorHandler(logger *slog.Logger) asynq.ErrorHandler {
	return asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
		retry, _ := asynq.GetRetryCount(ctx)

		logger.Error("task_failed",
			slog.String("type", t.Type()),
			slog.String("attempt", fmt.Sprintf("%d/%d", retry+1, maxTaskAttempts)),
			slog.String("error", err.Error()))
	})
}

type asynqSlogLogger struct {
	logger *slog.Logger
}

func newAsynqLogger(logger *slog.Logger) asynq.Logger {
	return &asynqSlogLogger{logger: logger}
}

func (l *asynqSlogLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqSlogLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqSlogLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqSlogLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqSlogLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
