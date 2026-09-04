package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fazal-khan/go_sync/internal/core"
	"github.com/fazal-khan/go_sync/internal/dbetls"
	"github.com/joho/godotenv"
)

func main() {
	// initialize the application
	appctx := Init()

	// load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file", err)
	}
	slog.Info("environment variables loaded successfully")

	if err := initServices(appctx); err != nil {
		slog.Error("failed to initialize services", "error", err)
		os.Exit(1)
	}

	appctx.Cron.Start()

	sig, cancel := waitChan()
	defer cancel()
	slog.Info("waiting for termination signal", "signal", sig)

	<-sig.Done()
	waitMore(appctx)

}

func initServices(appctx *core.AppCtx) error {

	appctx.Logger.Info("initializing services")
	// Initialize DBService
	dbService, err := dbetls.NewDBService(appctx)
	if err != nil {
		appctx.Logger.Error("failed to initialize DBService", slog.Any("error", err))
		return err
	}

	err = dbService.Init(appctx)
	if err != nil {
		appctx.Logger.Error("failed to initialize DBService", slog.Any("error", err))
		return err
	}
	return nil
}

func waitChan() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(
		context.Background(),
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // Docker/Kubernetes
	)
}

func waitMore(appctx *core.AppCtx) {
	// Give running work some time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	done := make(chan struct{})

	go func() {
		appctx.Cron.StopAndWait()
		time.Sleep(2 * time.Second)
		close(done)
	}()

	select {
	case <-done:
		appctx.Logger.Info("graceful shutdown completed")
		// Graceful shutdown completed.
	case <-shutdownCtx.Done():
		// Shutdown timeout exceeded.
	}
}
