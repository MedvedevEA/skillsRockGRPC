package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	auth "skillsRockGRPC/grpc/gen"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/httpserver"
	"skillsRockGRPC/internal/service"
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
	httpServer *httpserver.HttpServer
	cfg        *config.Grpc
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
func New(authServer *service.Service, httpServer *httpserver.HttpServer, lg *slog.Logger, cfg *config.Grpc) *GRPCServer {
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
	auth.RegisterAuthServiceServer(grpcServer, authServer)

	return &GRPCServer{
		lg:         lg,
		gRPCServer: grpcServer,
		httpServer: httpServer,
		cfg:        cfg,
	}
}
func (g *GRPCServer) Run() {
	chErr := make(chan error, 1)
	defer close(chErr)

	go func() {
		g.lg.Info("gRPC server start")
		listener, err := net.Listen("tcp", g.cfg.Addr)
		if err != nil {
			chErr <- err
			return
		}
		chErr <- g.gRPCServer.Serve(listener)
	}()
	go func() {
		chQuit := make(chan os.Signal, 1)
		signal.Notify(chQuit, syscall.SIGINT, syscall.SIGTERM)
		<-chQuit
		g.gRPCServer.GracefulStop()
		chErr <- nil
	}()
	if err := <-chErr; err != nil {
		g.lg.Error("gRPC server error", slog.Any("error", err))
		return
	}

	g.lg.Info("gRPC server stop")
}
