package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	pb "skillsRockGRPC/gen/go/auth/v3"
	"skillsRockGRPC/internal/authservice"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/pkg/servererrors"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	lg         *slog.Logger
	gRPCServer *grpc.Server
	cfg        *config.Api
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
func New(authServer *authservice.AuthService, lg *slog.Logger, cfg *config.Api) *GRPCServer {
	const op = "grpcserver.New"
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			lg.Error("Recovered from panic", slog.String("op", op), slog.Any("panic", p))
			return status.Error(codes.Internal, servererrors.ErrInternalServerError.Error())
		}),
	}
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(lg), loggingOpts...),
	))
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
	const op = "grpcserver.Stop"
	g.gRPCServer.GracefulStop()
	g.lg.Info(fmt.Sprintf("gRPC Server '%s' is stopped", g.cfg.Name), slog.String("op", op))
}
