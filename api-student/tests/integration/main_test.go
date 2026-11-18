package integration_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"backend-go/api-student/app/model"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test"))

	dsn := os.Getenv("TEST_DATABASE_URL")

	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL tidak diisi: integration test dilewati")
		os.Exit(0)
	}

	cfg, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}

	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("menolak berjalan: database %q tidak berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	cfg.MaxConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, cfg)

	if err == nil {
		err = pool.Ping(ctx)
	}

	if err != nil {
		fmt.Println("gagal terhubung ke database test:", err)
		os.Exit(1)
	}

	if err := applyMigrations(ctx); err != nil {
		fmt.Println("gagal menjalankan migration:", err)
		os.Exit(1)
	}

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.sql"))

	if err != nil {
		return err
	}

	slices.Sort(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)

		if err != nil {
			return err
		}

		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
	}

	return nil
}

func uniqueSuffix(t *testing.T) string {
	t.Helper()

	b := make([]byte, 4)

	if _, err := rand.Read(b); err != nil {
		t.Fatalf("membuat akhiran acak: %v", err)
	}

	return hex.EncodeToString(b)
}

func seedStudent(t *testing.T) model.User {
	t.Helper()

	u := model.User{
		Username: "mhs" + uniqueSuffix(t),
		Email:    "mhs" + uniqueSuffix(t) + "@student.unair.ac.id",
		Role:     "user",
		IsActive: true,
	}

	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, email, password, role, is_active)
		 VALUES ($1, $2, 'x', $3, $4)
		 RETURNING id`, u.Username, u.Email, u.Role, u.IsActive).Scan(&u.ID)

	if err != nil {
		t.Fatalf("menyimpan mahasiswa: %v", err)
	}

	return u
}

func seedCourse(t *testing.T, credits, quota int) model.Course {
	t.Helper()

	c := model.Course{
		Code:    "IF" + uniqueSuffix(t),
		Name:    "Kelas uji " + uniqueSuffix(t),
		Credits: credits,
		Quota:   quota,
	}

	err := pool.QueryRow(context.Background(),
		`INSERT INTO courses (code, name, credits, quota, enrolled_count)
		 VALUES ($1, $2, $3, $4, 0)
		 RETURNING id`, c.Code, c.Name, c.Credits, c.Quota).Scan(&c.ID)

	if err != nil {
		t.Fatalf("menyimpan kelas: %v", err)
	}

	return c
}
