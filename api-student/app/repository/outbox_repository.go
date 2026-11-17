package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"backend-go/api-student/app/model"
)

type OutboxRepository interface {
	WithTx(tx DBTX) OutboxRepository

	Insert(ctx context.Context, eventType string, event any) error
	FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
	Stats(ctx context.Context) (model.QueueStats, error)

	InsertEvent(ctx context.Context, db DBTX, eventType string, payload []byte) error
	FindUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	CreateNotification(ctx context.Context, db DBTX, n model.Notification) error
	MarkSent(ctx context.Context, outboxID int64) error
	QueueStatus(ctx context.Context) (map[string]int, error)
}

type outboxPostgresRepository struct {
	db DBTX
}

func NewOutboxRepository(db DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: db}
}

func (r *outboxPostgresRepository) WithTx(tx DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: tx}
}

func (r *outboxPostgresRepository) Insert(ctx context.Context, eventType string, event any) error {
	payload, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("menyusun payload outbox: %w", err)
	}

	_, err = r.db.Exec(ctx,
		"INSERT INTO outbox (event_type, payload) VALUES ($1, $2)", eventType, payload)

	if err != nil {
		return fmt.Errorf("menyimpan event outbox: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, event_type, payload, created_at
		 FROM outbox
		 WHERE published_at IS NULL
		 ORDER BY id
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`, limit)

	if err != nil {
		return nil, fmt.Errorf("mengambil outbox: %w", err)
	}

	defer rows.Close()

	result := []model.OutboxEvent{}

	for rows.Next() {
		var e model.OutboxEvent

		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca outbox: %w", err)
		}

		result = append(result, e)
	}

	return result, rows.Err()
}

func (r *outboxPostgresRepository) MarkPublished(ctx context.Context, ids []int64) error {
	_, err := r.db.Exec(ctx,
		"UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)", ids)

	if err != nil {
		return fmt.Errorf("menandai outbox terpublish: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) Stats(ctx context.Context) (model.QueueStats, error) {
	stats := model.QueueStats{}

	stmt := []struct {
		target *int
		sql    string
	}{
		{&stats.OutboxPending, "SELECT count(*) FROM outbox WHERE published_at IS NULL"},
		{&stats.OutboxTotal, "SELECT count(*) FROM outbox"},
		{&stats.PendingNotifications, "SELECT count(*) FROM notifications WHERE sent_at IS NULL"},
		{&stats.TotalNotifications, "SELECT count(*) FROM notifications"},
	}

	for _, item := range stmt {
		if err := r.db.QueryRow(ctx, item.sql).Scan(item.target); err != nil {
			return model.QueueStats{}, fmt.Errorf("menghitung antrean: %w", err)
		}
	}

	return stats, nil
}

func (r *outboxPostgresRepository) InsertEvent(ctx context.Context, db DBTX, eventType string, payload []byte) error {
	_, err := db.Exec(ctx,
		"INSERT INTO outbox (event_type, payload) VALUES ($1, $2)", eventType, payload)

	if err != nil {
		return fmt.Errorf("menyimpan event outbox: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) FindUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, event_type, payload, created_at, published_at
		 FROM outbox
		 WHERE published_at IS NULL
		 ORDER BY id
		 LIMIT $1`, limit)

	if err != nil {
		return nil, fmt.Errorf("mengambil outbox: %w", err)
	}

	defer rows.Close()

	result := []model.OutboxEvent{}

	for rows.Next() {
		var e model.OutboxEvent

		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.CreatedAt, &e.PublishedAt); err != nil {
			return nil, fmt.Errorf("membaca outbox: %w", err)
		}

		result = append(result, e)
	}

	return result, rows.Err()
}

func (r *outboxPostgresRepository) CreateNotification(ctx context.Context, db DBTX, n model.Notification) error {
	_, err := db.Exec(ctx,
		`INSERT INTO notifications (outbox_id, user_id, subject, body)
		 VALUES ($1, $2, $3, $4)`, n.OutboxID, n.UserID, n.Subject, n.Body)

	if err != nil {
		return fmt.Errorf("menyimpan notifikasi: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) MarkSent(ctx context.Context, outboxID int64) error {
	_, err := r.db.Exec(ctx,
		"UPDATE notifications SET sent_at = NOW() WHERE outbox_id = $1 AND sent_at IS NULL", outboxID)

	if err != nil {
		return fmt.Errorf("menandai notifikasi terkirim: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) QueueStatus(ctx context.Context) (map[string]int, error) {
	result := map[string]int{}

	stmt := []struct {
		key string
		sql string
	}{
		{"outbox_belum_publish", "SELECT count(*) FROM outbox WHERE published_at IS NULL"},
		{"outbox_total", "SELECT count(*) FROM outbox"},
		{"notifikasi_belum_kirim", "SELECT count(*) FROM notifications WHERE sent_at IS NULL"},
		{"notifikasi_total", "SELECT count(*) FROM notifications"},
	}

	for _, item := range stmt {
		var jumlah int

		if err := r.db.QueryRow(ctx, item.sql).Scan(&jumlah); err != nil {
			return nil, fmt.Errorf("menghitung antrean: %w", err)
		}

		result[item.key] = jumlah
	}

	return result, nil
}
