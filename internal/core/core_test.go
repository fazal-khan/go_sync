package core

import (
	"log/slog"
	"os"
	"testing"

	cron "github.com/netresearch/go-cron"
)

func TestAppCtx_New(t *testing.T) {
	ctx := &AppCtx{
		Logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})),
		Cron:   cron.New(),
	}
	if ctx.Logger == nil {
		t.Error("Logger should not be nil")
	}
	if ctx.Cron == nil {
		t.Error("Cron should not be nil")
	}
}

func TestAppCtx_NilFields(t *testing.T) {
	ctx := &AppCtx{}
	if ctx.Logger != nil {
		t.Error("Logger should be nil by default")
	}
	if ctx.Cron != nil {
		t.Error("Cron should be nil by default")
	}
}
