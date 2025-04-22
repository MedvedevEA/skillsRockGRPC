package main

import (
	"context"
	"os"
	"os/signal"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/grpcserver"
	"skillsRockGRPC/internal/logger"
	"skillsRockGRPC/internal/service"
	"skillsRockGRPC/internal/store"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	lg := logger.MustNew(cfg.Env)

	store := store.MustNew(context.Background(), lg, &cfg.PostgreSQL)

	service := service.MustNew(store, lg, cfg.Token.SecretPath)

	grpcServer := grpcserver.New(service, lg, &cfg.Api)
	go func() {
		chQuit := make(chan os.Signal, 1)
		signal.Notify(chQuit, syscall.SIGINT, syscall.SIGTERM)
		<-chQuit
		grpcServer.Stop()

	}()
	grpcServer.Run()
}
