package authservice

import (
	"context"
	"log/slog"
	pb "skillsRockGRPC/gen/go/auth/v1"
	srvDto "skillsRockGRPC/internal/service/dto"
)

type Service interface {
	Register(dto *srvDto.Register) (string, error)
	Login(dto *srvDto.Login) (string, error)
	CheckToken(dto *srvDto.CheckToken) (bool, error)
}

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	service Service
	lg      *slog.Logger
}

func New(service Service, lg *slog.Logger) *AuthService {
	return &AuthService{
		service: service,
		lg:      lg,
	}
}

func (a *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	message, err := a.service.Register(&srvDto.Register{
		Login:    req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{Message: message}, nil
}
func (a *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := a.service.Login(&srvDto.Login{
		Login:    req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: token}, nil
}
func (a *AuthService) CheckToken(ctx context.Context, req *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	ok, err := a.service.CheckToken(&srvDto.CheckToken{
		Token: req.Token,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CheckTokenResponse{Ok: ok}, nil
}
