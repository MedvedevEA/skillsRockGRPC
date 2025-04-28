package grpcserver

import (
	"fmt"
	"log/slog"
	"net"

	pb "skillsRockGRPC/gen/go/auth/v3"
	"skillsRockGRPC/internal/authservice"
	"skillsRockGRPC/internal/config"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	lg         *slog.Logger
	gRPCServer *grpc.Server
	cfg        *config.Api
}

func New(authServer *authservice.AuthService, lg *slog.Logger, cfg *config.Api) *GRPCServer {
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	return &GRPCServer{
		lg:         lg,
		gRPCServer: grpcServer,
		cfg:        cfg,
	}
}
func (g *GRPCServer) Run() error {
	const op = "grpcserver.Run"
	listener, err := net.Listen("tcp", g.cfg.Addr)
	if err != nil {
		g.lg.Error("application error", slog.String("op", op), slog.Any("error", err))
	}
	g.lg.Info(fmt.Sprintf("gRPC Server '%s' is started in addr:[%s]", g.cfg.Name, g.cfg.Addr), slog.String("op", op))
	if err := g.gRPCServer.Serve(listener); err != nil {
		g.lg.Error("application error", slog.String("op", op), slog.Any("error", err))
		return err
	}
	return nil
}
func (g *GRPCServer) Stop() {
	const op = "grpcapp.Stop"
	g.gRPCServer.GracefulStop()
	g.lg.Info(fmt.Sprintf("gRPC Server '%s' is stopped", g.cfg.Name), slog.String("op", op))
}
