package core

import (
	"log/slog"

	"github.com/fazal-khan/go_sync/internal/util"
)

type AppCtx struct {
	Logger *slog.Logger
}

func (ctx *AppCtx) Getenv(key, v string) string {
	return util.Getenv(key, v)
}
