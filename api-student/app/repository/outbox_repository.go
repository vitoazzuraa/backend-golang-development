package repository

import (
	"context"
	"fmt"

	"backend-go/api-student/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository interface {
	InsertEvent(ctx context.Context, db DBTX, eventType string, payload []byte) error
	FindUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
	CreateNotification(ctx context.Context, db DBTX, n model.Notification) error
	MarkSent(ctx context.Context, outboxID int64) error
	QueueStatus(ctx context.Context) (map[string]int, error)
}

type outboxPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) OutboxRepository {
	return &outboxPostgresRepository{pool: pool}
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
	rows, err := r.pool.Query(ctx,
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

func (r *outboxPostgresRepository) MarkPublished(ctx context.Context, ids []int64) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)", ids)

	if err != nil {
		return fmt.Errorf("menandai outbox terpublish: %w", err)
	}

	return nil
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
	_, err := r.pool.Exec(ctx,
		"UPDATE notifications SET sent_at = NOW() WHERE outbox_id = $1 AND sent_at IS NULL", outboxID)

	if err != nil {
		return fmt.Errorf("menandai notifikasi terkirim: %w", err)
	}

	return nil
}

func (r *outboxPostgresRepository) QueueStatus(ctx context.Context) (map[string]int, error) {
	hasil := map[string]int{}

	pernyataan := []struct {
		kunci string
		sql   string
	}{
		{"outbox_belum_publish", "SELECT count(*) FROM outbox WHERE published_at IS NULL"},
		{"outbox_total", "SELECT count(*) FROM outbox"},
		{"notifikasi_belum_kirim", "SELECT count(*) FROM notifications WHERE sent_at IS NULL"},
		{"notifikasi_total", "SELECT count(*) FROM notifications"},
	}

	for _, item := range pernyataan {
		var jumlah int

		if err := r.pool.QueryRow(ctx, item.sql).Scan(&jumlah); err != nil {
			return nil, fmt.Errorf("menghitung antrean: %w", err)
		}

		hasil[item.kunci] = jumlah
	}

	return hasil, nil
}
