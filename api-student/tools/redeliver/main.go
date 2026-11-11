package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"backend-go/api-student/app/model"
	"backend-go/api-student/config"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	outboxID = flag.Int("outbox", 0, "ID event outbox yang akan dikirim ulang")
	copies   = flag.Int("copies", 1, "Jumlah salinan task yang dikirim ke Redis")
)

type outboxEvent struct {
	EventType string
	Payload   string
}

func main() {
	config.LoadEnv()
	flag.Parse()

	if *outboxID < 1 {
		fmt.Println("flag -outbox wajib diisi dengan id event yang valid")
		os.Exit(1)
	}

	if *copies < 1 {
		fmt.Println("flag -copies harus lebih besar dari nol")
		os.Exit(1)
	}

	event, err := loadEvent(*outboxID)

	if err != nil {
		fmt.Println("gagal membaca event outbox:", err)
		os.Exit(1)
	}

	client := asynq.NewClient(config.RedisOpt())

	defer client.Close()

	for i := 1; i <= *copies; i++ {
		taskID := fmt.Sprintf("outbox-%d-dup-%d", *outboxID, i)

		if err := enqueue(client, taskID, *outboxID, event); err != nil {
			fmt.Printf("salinan %d/%d gagal dikirim: %v\n", i, *copies, err)
			os.Exit(1)
		}

		fmt.Printf("salinan %d/%d dikirim dengan task id %s\n", i, *copies, taskID)
	}
}

func loadEvent(id int) (outboxEvent, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "go_backend"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return outboxEvent{}, err
	}

	defer pool.Close()

	var event outboxEvent

	err = pool.QueryRow(ctx,
		"SELECT event_type, payload FROM outbox WHERE id = $1", id).Scan(&event.EventType, &event.Payload)

	if err != nil {
		return outboxEvent{}, err
	}

	return event, nil
}

func enqueue(client *asynq.Client, taskID string, id int, event outboxEvent) error {
	job := model.NotificationJob{
		OutboxID:  int64(id),
		EventType: event.EventType,
		Payload:   json.RawMessage(event.Payload),
	}

	payload, err := json.Marshal(job)

	if err != nil {
		return err
	}

	task := asynq.NewTask(model.TaskSendNotification, payload,
		asynq.TaskID(taskID),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour))

	_, err = client.EnqueueContext(context.Background(), task)

	return err
}
