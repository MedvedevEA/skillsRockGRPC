package main

import (
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/grpcserver"
	"skillsRockGRPC/internal/httpserver"
	"skillsRockGRPC/internal/logger"
	"skillsRockGRPC/internal/scheduler"
	"skillsRockGRPC/internal/service"
	"skillsRockGRPC/internal/store"
)

func main() {
	cfg := config.MustLoad()

	lg := logger.MustNew(cfg.Env)

	store := store.MustNew(lg, &cfg.Store)

	service := service.MustNew(store, lg, &cfg.Token)

	httpServer := httpserver.MustNew(lg, &cfg.Http, &cfg.Grpc)
	httpServer.Run()

	grpcServer := grpcserver.New(service, lg, &cfg.Grpc)

	scheduler := scheduler.New(lg, &cfg.Scheduler)
	scheduler.RemoveRefreshTokens(store.RemoveRefreshTokensByExpirationAt)

	grpcServer.Run()

	httpServer.Stop()
	scheduler.Stop()

}
