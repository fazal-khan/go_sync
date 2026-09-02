package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// creates a new logger, enhance method to accept options for configuration.
func New(fileName string, maxSizeMB int, maxBackups int, maxAgeDays int) *slog.Logger {
	rollingFile := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   true, // disabled by default
	}

	// Log to both File and Console
	multiWriter := io.MultiWriter(os.Stdout, rollingFile)

	// Create JSON handler for slog
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		AddSource: true, // Optional: Includes the file name and line number
		Level:     slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func NewConsoleLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true, // Optional: Includes the file name and line number
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
