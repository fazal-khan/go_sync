package core

import (
	"log/slog"

	"github.com/fazal-khan/go_sync/internal/util"
	cron "github.com/netresearch/go-cron"
)

type AppCtx struct {
	Logger *slog.Logger
	Cron   *cron.Cron
}

func (ctx *AppCtx) Getenv(key, v string) string {
	return util.Getenv(key, v)
}
