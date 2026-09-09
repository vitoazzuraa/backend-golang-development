package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() *slog.Logger {
	_, filename, _, _ := runtime.Caller(0)

	base := filepath.Dir(filepath.Dir(filename))

	logPath := filepath.Join(base, "logs", "app.log")

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		panic("gagal membuat folder logs: " + err.Error())
	}

	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     14,
		Compress:   true,
	}

	writer := io.MultiWriter(os.Stdout, rotator)

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(GetEnv("LOG_LEVEL", "info")),
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
	
	return logger
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
