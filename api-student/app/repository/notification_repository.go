package repository

import (
	"context"
	"errors"
	"fmt"

	"backend-go/api-student/app/model"

	"github.com/jackc/pgx/v5"
)

type NotificationRepository interface {
	WithTx(tx DBTX) NotificationRepository

	Claim(ctx context.Context, outboxID int64, userID int, subject, body string) (bool, error)
	MarkSent(ctx context.Context, outboxID int64, userID int) error
	FindByUser(ctx context.Context, userID int) ([]model.Notification, error)
	Count(ctx context.Context, outboxID int64) (int, error)
	FindTargets(ctx context.Context, outboxID int64) ([]int, error)
}

type notificationPostgresRepository struct {
	db DBTX
}

func NewNotificationRepository(db DBTX) NotificationRepository {
	return &notificationPostgresRepository{db: db}
}

func (r *notificationPostgresRepository) WithTx(tx DBTX) NotificationRepository {
	return &notificationPostgresRepository{db: tx}
}

func (r *notificationPostgresRepository) Claim(ctx context.Context, outboxID int64, userID int, subject, body string) (bool, error) {
	var id int64

	err := r.db.QueryRow(ctx,
		`INSERT INTO notifications (outbox_id, user_id, subject, body)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (outbox_id, user_id) DO NOTHING
		 RETURNING id`, outboxID, userID, subject, body).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("mengklaim notifikasi: %w", err)
	}

	return true, nil
}

func (r *notificationPostgresRepository) MarkSent(ctx context.Context, outboxID int64, userID int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE notifications SET sent_at = NOW() WHERE outbox_id = $1 AND user_id = $2", outboxID, userID)

	if err != nil {
		return fmt.Errorf("menandai notifikasi terkirim: %w", err)
	}

	return nil
}

func (r *notificationPostgresRepository) FindByUser(ctx context.Context, userID int) ([]model.Notification, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, outbox_id, user_id, subject, body, sent_at
		 FROM notifications
		 WHERE user_id = $1
		 ORDER BY outbox_id DESC`, userID)

	if err != nil {
		return nil, fmt.Errorf("mengambil notifikasi: %w", err)
	}

	defer rows.Close()

	result := []model.Notification{}

	for rows.Next() {
		var n model.Notification

		if err := rows.Scan(&n.ID, &n.OutboxID, &n.UserID, &n.Subject, &n.Body, &n.SentAt); err != nil {
			return nil, fmt.Errorf("membaca notifikasi: %w", err)
		}

		result = append(result, n)
	}

	return result, rows.Err()
}

func (r *notificationPostgresRepository) Count(ctx context.Context, outboxID int64) (int, error) {
	var total int

	err := r.db.QueryRow(ctx,
		"SELECT count(*) FROM notifications WHERE outbox_id = $1", outboxID).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("menghitung notifikasi: %w", err)
	}

	return total, nil
}

func (r *notificationPostgresRepository) FindTargets(ctx context.Context, outboxID int64) ([]int, error) {
	rows, err := r.db.Query(ctx,
		"SELECT user_id FROM quota_notification_targets WHERE outbox_id = $1 ORDER BY user_id", outboxID)

	if err != nil {
		return nil, fmt.Errorf("mengambil tujuan notifikasi: %w", err)
	}

	defer rows.Close()

	result := []int{}

	for rows.Next() {
		var userID int

		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("membaca tujuan notifikasi: %w", err)
		}

		result = append(result, userID)
	}

	return result, rows.Err()
}
