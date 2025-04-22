package grpcserver

import (
	"fmt"
	"log/slog"
	"net"

	pb "skillsRockGRPC/gen/go/auth/v1"
	"skillsRockGRPC/internal/authservice"
	"skillsRockGRPC/internal/config"

	"google.golang.org/grpc"
)

type App struct {
	lg         *slog.Logger
	gRPCServer *grpc.Server
	cfg        *config.Api
}

func New(service authservice.Service, lg *slog.Logger, cfg *config.Api) *App {
	grpcServer := grpc.NewServer()
	authServer := authservice.New(service, lg)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	return &App{
		lg:         lg,
		gRPCServer: grpcServer,
		cfg:        cfg,
	}
}
func (a *App) Run() error {
	/*
			listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("gRPC сервер запущен на :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal(err)
		}

	*/
	const op = "app.Run"
	listener, err := net.Listen("tcp", a.cfg.Addr)
	if err != nil {
		a.lg.Error("application error", slog.String("op", op), slog.Any("error", err))
	}
	a.lg.Info(fmt.Sprintf("gRPC Server '%s' is started in addr:[%s]", a.cfg.Name, a.cfg.Addr), slog.String("op", op))
	if err := a.gRPCServer.Serve(listener); err != nil {
		a.lg.Error("application error", slog.String("op", op), slog.Any("error", err))
		return err
	}
	return nil
}
func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.gRPCServer.GracefulStop()
	a.lg.Info(fmt.Sprintf("gRPC Server '%s' is stopped", a.cfg.Name), slog.String("op", op))
}
