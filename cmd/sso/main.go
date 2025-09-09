package main

import (
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/lib/logger/setuplogger"
	"syscall"

	"golang.org/x/exp/slog"
)

func main() {
	cfg := config.MustLoad()
	log := setuplogger.Setup(cfg.Env)
	log.Info("starting application")

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TlsPath, cfg.TokenTTL)
	go application.GRPCSrv.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))
	application.GRPCSrv.Stop()
	log.Info("application stopped")
}
