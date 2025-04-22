package main

import (
	"context"
	"os"
	"os/signal"
	"skillsRockGRPC/internal/authservice"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/grpcserver"
	"skillsRockGRPC/internal/logger"
	"skillsRockGRPC/internal/store"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	lg := logger.MustNew(cfg.Env)

	store := store.MustNew(context.Background(), lg, &cfg.PostgreSQL)

	authService := authservice.New(store, lg, &cfg.Token)

	grpcServer := grpcserver.New(authService, lg, &cfg.Api)
	go func() {
		chQuit := make(chan os.Signal, 1)
		signal.Notify(chQuit, syscall.SIGINT, syscall.SIGTERM)
		<-chQuit
		grpcServer.Stop()

	}()
	grpcServer.Run()
}
