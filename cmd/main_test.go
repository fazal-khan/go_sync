package main

import (
	"testing"

	"github.com/fazal-khan/go_sync/internal/logger"
	cron "github.com/netresearch/go-cron"
)

func TestInit(t *testing.T) {
	ctx := Init()
	if ctx == nil {
		t.Fatal("Init returned nil")
	}
	if ctx.Logger == nil {
		t.Error("Logger should not be nil")
	}
	if ctx.Cron == nil {
		t.Error("Cron should not be nil")
	}
	_ = cron.New()
}

func TestInit_Logger(t *testing.T) {
	log := logger.NewConsoleLogger()
	if log == nil {
		t.Fatal("NewConsoleLogger returned nil")
	}
	log.Info("test message")
}