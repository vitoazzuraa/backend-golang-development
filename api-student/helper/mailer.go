package helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"backend-go/api-student/app/model"
)

var (
	ErrTemporaryDelivery = errors.New("pengiriman gagal sementara")
	ErrPermanentDelivery = errors.New("pengiriman gagal permanen")
)

type Mailer interface {
	Send(ctx context.Context, msg model.Message) error
}

type LogMailer struct {
	mu        sync.Mutex
	path      string
	delay     time.Duration
	failEvery int64
	attempts  atomic.Int64
}

func NewLogMailer(path string, delay time.Duration, failEvery int) (*LogMailer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("menyiapkan folder log email: %w", err)
	}

	return &LogMailer{path: path, delay: delay, failEvery: int64(failEvery)}, nil
}

func (m *LogMailer) Send(ctx context.Context, msg model.Message) error {
	n := m.attempts.Add(1)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.delay):
	}

	if strings.HasSuffix(msg.To, ".invalid") {
		return fmt.Errorf("%w: alamat %s tidak valid", ErrPermanentDelivery, msg.To)
	}

	if m.failEvery > 0 && n%m.failEvery == 0 {
		return fmt.Errorf("%w: server email sibuk", ErrTemporaryDelivery)
	}

	m.mu.Lock()

	defer m.mu.Unlock()

	file, err := os.OpenFile(m.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("%w: membuka log email: %v", ErrTemporaryDelivery, err)
	}

	defer file.Close()

	if _, err := fmt.Fprintf(file, "%s to=%s subject=%q\n",
		time.Now().Format(time.RFC3339), msg.To, msg.Subject); err != nil {
		return fmt.Errorf("%w: menulis log email: %v", ErrTemporaryDelivery, err)
	}

	return nil
}
