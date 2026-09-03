package main

import (
	"log/slog"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/logger"
	cron "github.com/netresearch/go-cron"
)

func Init() *core.AppCtx {
	log := initLogger()

	cronlog := cron.NewSlogLogger(log)
	cronoption := cron.WithChain(cron.Recover(cronlog), cron.SkipIfStillRunning(cronlog))

	c := cron.New(cronoption)

	ctx := core.AppCtx{
		Logger: log,
		Cron:   c,
	}

	return &ctx
}

func initLogger() *slog.Logger {
	logger := logger.NewConsoleLogger()
	logger.Info("in main, logger initialized")
	return logger
}
