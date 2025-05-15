package httpserver

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"skillsRockGRPC/internal/config"

	auth "skillsRockGRPC/grpc/gen"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HttpServer struct {
	lg         *slog.Logger
	httpServer *http.Server
	cfg        *config.Http
}

func MustNew(lg *slog.Logger, cfgHttp *config.Http, cfgGrpc *config.Grpc) *HttpServer {

	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, cfgGrpc.Addr, opts)
	if err != nil {
		log.Fatalf("HTTP server: %v", err)
	}
	httpServer := &http.Server{
		Addr:    cfgHttp.Addr,
		Handler: mux,
	}

	return &HttpServer{
		lg:         lg,
		httpServer: httpServer,
		cfg:        cfgHttp,
	}
}
func (h *HttpServer) Run() {
	go func() {
		h.lg.Info("HTTP serever start", slog.String("addr", h.cfg.Addr))
		if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.lg.Error("HTTP server error", slog.Any("error", err))
		}
	}()
}
func (h *HttpServer) Stop() {
	err := h.httpServer.Shutdown(context.Background())
	if err != nil {
		h.lg.Error("HTTP server error", slog.Any("error", err))
		return
	}
	h.lg.Info("HTTP server stop")
}
