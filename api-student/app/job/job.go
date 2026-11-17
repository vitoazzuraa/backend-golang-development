package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"

	"github.com/hibiken/asynq"
)

const TaskSendNotification = "notifikasi:kirim"

type Mailer struct {
	delay     time.Duration
	failEvery int
	mu        sync.Mutex
	attempt   int
	logPath   string
}

func NewMailer(logPath string) *Mailer {
	delay := time.Duration(envInt("MAILER_DELAY_MS", 1500)) * time.Millisecond

	return &Mailer{
		delay:     delay,
		failEvery: envInt("MAILER_FAIL_EVERY", 0),
		logPath:   logPath,
	}
}

func (m *Mailer) Send(subject, body string) error {
	time.Sleep(m.delay)

	m.mu.Lock()

	defer m.mu.Unlock()

	m.attempt++

	if m.failEvery > 0 && m.attempt%m.failEvery == 0 {
		return fmt.Errorf("pengiriman email gagal sementara")
	}

	file, err := os.OpenFile(m.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%s | %s | %s\n", time.Now().Format(time.RFC3339), subject, body)

	return err
}

type Relay struct {
	outbox   repository.OutboxRepository
	client   *asynq.Client
	interval time.Duration
	batch    int
}

func NewRelay(outbox repository.OutboxRepository, client *asynq.Client) *Relay {
	return &Relay{
		outbox:   outbox,
		client:   client,
		interval: time.Duration(envInt("RELAY_INTERVAL_MS", 1000)) * time.Millisecond,
		batch:    envInt("RELAY_BATCH", 100),
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.publishOnce(ctx)
		}
	}
}

func (r *Relay) publishOnce(ctx context.Context) {
	events, err := r.outbox.FindUnpublished(ctx, r.batch)

	if err != nil || len(events) == 0 {
		return
	}

	ids := make([]int64, 0, len(events))

	for _, event := range events {
		task := asynq.NewTask(TaskSendNotification, []byte(event.Payload),
			asynq.TaskID(strconv.FormatInt(event.ID, 10)),
			asynq.MaxRetry(5))

		if _, err := r.client.EnqueueContext(ctx, task); err != nil {
			if err == asynq.ErrTaskIDConflict {
				ids = append(ids, event.ID)
			}

			continue
		}

		ids = append(ids, event.ID)
	}

	if len(ids) > 0 {
		_ = r.outbox.MarkPublished(ctx, ids)
	}
}

type Worker struct {
	outbox repository.OutboxRepository
	mailer *Mailer
}

func NewWorker(outbox repository.OutboxRepository, mailer *Mailer) *Worker {
	return &Worker{outbox: outbox, mailer: mailer}
}

func (w *Worker) Handler(ctx context.Context, t *asynq.Task) error {
	var event model.OutboxEvent

	event.Payload = string(t.Payload())

	if err := w.mailer.Send("notifikasi", event.Payload); err != nil {
		return err
	}

	id, err := strconv.ParseInt(t.ResultWriter().TaskID(), 10, 64)

	if err == nil {
		_ = w.outbox.MarkSent(ctx, id)
	}

	return nil
}

func NewRedisOpt() asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{Addr: envString("REDIS_ADDR", "localhost:6379")}

	if db := envInt("REDIS_DB", 0); db > 0 {
		opt.DB = db
	}

	if password := envString("REDIS_PASSWORD", ""); password != "" {
		opt.Password = password
	}

	return opt
}

func NewServer() *asynq.Server {
	return asynq.NewServer(NewRedisOpt(), asynq.Config{
		Concurrency: envInt("WORKER_CONCURRENCY", 5),
	})
}

func LogPath() string {
	path := filepath.Join("api-student", "logs", "mail.log")

	_ = os.MkdirAll(filepath.Dir(path), 0755)

	return path
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)

	if err != nil {
		return fallback
	}

	return parsed
}
