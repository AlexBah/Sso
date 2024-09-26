package grpcapp

import (
	"log/slog"
	grpcapp "sso/internal/app"
	"time"
)

type App struct {
	gRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	// Инициализировать хранилище

	// init auth service

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
