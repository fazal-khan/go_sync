package main

import (
	"log/slog"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/logger"
)

func Init() {
	ctx := core.AppCtx{
		Logger: initLogger(),
	}

	initConfig()

	ctx.Logger.Info("in main, logger initialized")
}

func initLogger() *slog.Logger {
	return logger.NewConsoleLogger()
}

func initConfig() {

}
