package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/hibiken/asynq"
)

type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users:         users,
		notifications: notifications,
		tx:            tx,
		mailer:        mailer,
		logger:        logger,
	}
}

func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob

	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	var event model.EnrollmentEvent

	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event rusak: %v", asynq.SkipRetry, err)
	}

	user, err := w.users.FindByID(ctx, event.UserID)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
		}

		return err
	}

	msg, err := BuildMessage(event, job.EventType)

	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	msg.To = user.Email

	alreadySent := false

	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, job.OutboxID, user.ID, msg.Subject, msg.Body)

		if err != nil {
			return err
		}

		if !claimed {
			alreadySent = true

			return nil
		}

		return w.mailer.Send(ctx, msg)
	})

	if err != nil {
		if errors.Is(err, helper.ErrPermanentDelivery) {
			return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
		}

		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)

	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))

	return nil
}
