package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "skillsRockGRPC/gen/go/auth/v1"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/logger"

	"google.golang.org/grpc"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	fmt.Println(req.Username, req.Password)
	return &pb.RegisterResponse{
		Message: "TODO",
	}, nil
}
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{
		Token: "TODO",
	}, nil
}
func (s *AuthServer) CheckToken(ctx context.Context, req *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	return &pb.CheckTokenResponse{
		Ok: false,
	}, nil

}

func main() {
	cfg := config.MustLoad()

	lg := logger.MustNew(cfg.Env)

	store := postgresql.MustNew(context.Background(), lg, &cfg.PostgreSQL)

	service := service.New(store, lg)

	apiServer := apiserver.New(service, lg, &cfg.Api)
	apiServer.MustRun()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &AuthServer{})
	log.Printf("gRPC сервер запущен на :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
