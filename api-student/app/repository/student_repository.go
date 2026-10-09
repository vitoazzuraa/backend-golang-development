package repository

import (
	"context"
	"errors"
	"fmt"

	"backend-go/api-student/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
	Update(ctx context.Context, student model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func (r *studentPostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error) {
	where, args := buildCursorFilter(q)

	args = append(args, q.Limit+1)

	rows, err := r.pool.Query(ctx, `
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students`+where+fmt.Sprintf(
		" ORDER BY created_at DESC, id DESC LIMIT $%d",
		len(args),
	), args...)

	if err != nil {
		return nil, fmt.Errorf("mengambil daftar student: %w", err)
	}

	defer rows.Close()

	result := []model.Student{}

	for rows.Next() {
		var student model.Student
		if err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.OwnerID,
			&student.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca baris student: %w", err)
		}

		result = append(result, student)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query student: %w", err)
	}

	return result, nil
}

func buildCursorFilter(q model.CursorQuery) (string, []any) {
	where := " WHERE 1 = 1"

	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}
	if q.After != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)+1, len(args)+2)
		args = append(args, q.After.CreatedAt, q.After.ID)
	}

	return where, args
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var student model.Student

	err := r.pool.QueryRow(ctx, `
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students
		WHERE id = $1`,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}

		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}

	return student, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, student model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO students (nim, name, grade, is_active, owner_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.OwnerID,
	).Scan(
		&student.ID,
		&student.CreatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}

	return student, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, student model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE students
		SET nim = $1, name = $2, grade = $3, is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, owner_id, created_at`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.ID,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}

	return student, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)

	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
