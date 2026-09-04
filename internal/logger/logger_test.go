package logger

import (
	"os"
	"testing"
)

func TestNewConsoleLogger(t *testing.T) {
	log := NewConsoleLogger()
	if log == nil {
		t.Fatal("NewConsoleLogger returned nil")
	}
	log.Info("test message")
}

func TestNew_WithFile(t *testing.T) {
	dir := t.TempDir()
	logFile := dir + "/test.log"
	log := New(logFile, 10, 1, 1)
	if log == nil {
		t.Fatal("New returned nil")
	}
	log.Info("test message")
	log.Info("second message")

	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("log file should exist: %v", err)
	}
}

func TestNewConsoleLogger_Levels(t *testing.T) {
	log := NewConsoleLogger()
	if log == nil {
		t.Fatal("NewConsoleLogger returned nil")
	}
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")
}