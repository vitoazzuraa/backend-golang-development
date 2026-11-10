package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"backend-go/api-student/app/model"
	"backend-go/api-student/config"
	"backend-go/api-student/helper"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	baseURL      = flag.String("url", "http://localhost:3000/api/v1", "Base URL API")
	supervisorID = flag.Int("supervisor", 1, "Supervisor ID")
	target       = flag.Int("target", 30, "Jumlah mahasiswa penyerang")
)

func main() {
	config.LoadEnv()
	flag.Parse()

	fmt.Printf("%d mahasiswa berebut pembimbing ID %d secara serentak...\n", *target, *supervisorID)
	start := time.Now()

	ids := getAccounts(*target)
	if len(ids) < *target {
		fmt.Printf("PERINGATAN: cuma %d mahasiswa ditemukan, butuh %d\n", len(ids), *target)
		*target = len(ids)
	}

	tally := fire(*target, func(i int) string {
		token := makeToken(ids[i])
		status, _ := assign(token, *supervisorID, ids[i].ID)
		return fmt.Sprintf("%d %s", status, http.StatusText(status))
	})

	fmt.Printf("hasil request:\n")
	for k, v := range tally {
		fmt.Printf("  %-30s %d\n", k, v)
	}
	fmt.Printf("waktu total : %v\n", time.Since(start))
}

func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results <- job(i)
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

type account struct {
	ID       int64
	Username string
}

func getAccounts(limit int) []account {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "go_backend"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, "SELECT id, username FROM users WHERE username LIKE 'mhs%' ORDER BY id LIMIT $1", limit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	result := []account{}
	for rows.Next() {
		var a account
		rows.Scan(&a.ID, &a.Username)
		result = append(result, a)
	}
	return result
}

func makeToken(a account) string {
	secret := config.GetEnv("JWT_SECRET", "")
	issuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")
	mgr := helper.NewJWTManager(secret, issuer, 15*time.Minute)

	token, _ := mgr.GenerateAccess(model.User{
		ID:       int(a.ID),
		Username: a.Username,
		Role:     "admin",
	})
	return token
}

func assign(token string, supervisorID int, studentID int64) (int, string) {
	url := fmt.Sprintf("%s/supervisors/%d/assign", *baseURL, supervisorID)
	body := fmt.Sprintf(`{"student_id":%d}`, studentID)

	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}
