package main

import (
	"skillsRockGRPC/internal/authservice"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/grpcserver"
	"skillsRockGRPC/internal/logger"
	"skillsRockGRPC/internal/scheduler"
	"skillsRockGRPC/internal/store"
)

func main() {
	cfg := config.MustLoad()

	lg := logger.MustNew(cfg.Env)

	store := store.MustNew(lg, &cfg.Store)

	authService := authservice.MustNew(store, lg, &cfg.Token)

	grpcServer := grpcserver.New(authService, lg, &cfg.Api)

	scheduler := scheduler.New(lg, &cfg.Scheduler)
	scheduler.RemoveRefreshTokens(store.RemoveRefreshTokensByExpirationAt)

	grpcServer.Run()

	scheduler.Stop()

}
