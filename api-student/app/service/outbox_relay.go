package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"

	"github.com/hibiken/asynq"
)

type TaskEnqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type OutboxRelay struct {
	outbox        repository.OutboxRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	queue         TaskEnqueuer
	logger        *slog.Logger
	interval      time.Duration
	batch         int
}

func NewOutboxRelay(
	outbox repository.OutboxRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	queue TaskEnqueuer,
	logger *slog.Logger,
	interval time.Duration,
	batch int,
) *OutboxRelay {
	return &OutboxRelay{
		outbox:        outbox,
		notifications: notifications,
		tx:            tx,
		queue:         queue,
		logger:        logger,
		interval:      interval,
		batch:         batch,
	}
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		published, err := r.RelayOnce(ctx)

		if err != nil {
			r.logger.Error("relay_delayed",
				slog.Int("published", published),
				slog.String("error", err.Error()))
		} else {
			r.logger.Info("relay_done", slog.Int("published", published))
		}

		if err == nil && published == r.batch {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *OutboxRelay) RelayOnce(ctx context.Context) (int, error) {
	published := 0

	var enqueueErr error

	err := r.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		repo := r.outbox.WithTx(tx)

		events, err := repo.FetchUnpublished(ctx, r.batch)

		if err != nil {
			return err
		}

		ids := []int64{}

		for _, e := range events {
			if err := r.publishEvent(ctx, e); err != nil {
				enqueueErr = err
				break
			}

			ids = append(ids, e.ID)
		}

		if len(ids) == 0 {
			return nil
		}

		if err := repo.MarkPublished(ctx, ids); err != nil {
			return err
		}

		published = len(ids)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return published, enqueueErr
}

func (r *OutboxRelay) publishEvent(ctx context.Context, e model.OutboxEvent) error {
	if e.EventType != model.EventQuotaChanged {
		return r.publish(ctx, e, 0, fmt.Sprintf("outbox-%d", e.ID))
	}

	targets, err := r.notifications.FindTargets(ctx, e.ID)

	if err != nil {
		return err
	}

	for _, userID := range targets {
		taskID := fmt.Sprintf("outbox-%d-user-%d", e.ID, userID)

		if err := r.publish(ctx, e, userID, taskID); err != nil {
			return err
		}
	}

	return nil
}

func (r *OutboxRelay) publish(ctx context.Context, e model.OutboxEvent, userID int, taskID string) error {
	payload := json.RawMessage(e.Payload)

	if userID > 0 {
		var event model.EnrollmentEvent

		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("membaca payload event: %w", err)
		}

		event.UserID = userID

		rewritten, err := json.Marshal(event)

		if err != nil {
			return fmt.Errorf("menyusun ulang payload event: %w", err)
		}

		payload = rewritten
	}

	job := model.NotificationJob{OutboxID: e.ID, EventType: e.EventType, Payload: payload}

	body, err := json.Marshal(job)

	if err != nil {
		return fmt.Errorf("menyusun payload task: %w", err)
	}

	task := asynq.NewTask(model.TaskSendNotification, body,
		asynq.TaskID(taskID),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour))

	if _, err := r.queue.EnqueueContext(ctx, task); err != nil {
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			return nil
		}

		return fmt.Errorf("mengirim task ke antrean: %w", err)
	}

	return nil
}
