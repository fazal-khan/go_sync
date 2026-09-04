package core

import (
	"log/slog"

	cron "github.com/netresearch/go-cron"
)

type AppCtx struct {
	Logger *slog.Logger
	Cron   *cron.Cron
}
