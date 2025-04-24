package authservice

import (
	"context"
	"log"
	"log/slog"
	"os"
	pb "skillsRockGRPC/gen/go/auth/v2"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/repository"
	"skillsRockGRPC/internal/repository/dto"

	"skillsRockGRPC/pkg/jwttoken"
	"skillsRockGRPC/pkg/secret"
	"skillsRockGRPC/pkg/servererrors"
	"time"

	"github.com/pkg/errors"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	store           repository.Repository
	secretKey       []byte
	accessLifetime  time.Duration
	refrashLifetime time.Duration
	lg              *slog.Logger
}

func MustNew(store repository.Repository, lg *slog.Logger, cfg *config.Token) *AuthService {
	const op = "authservice.MustNew"
	secretByteArray, err := os.ReadFile(cfg.SecretPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	secretRSA, err := secret.UnmarshalRSAPrivate(secretByteArray)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	secretKey := secret.MarshalRSAPrivate(secretRSA, "")

	return &AuthService{
		store:           store,
		secretKey:       secretKey,
		accessLifetime:  cfg.AccessLifetime,
		refrashLifetime: cfg.RefreshLifetime,
		lg:              lg,
	}
}

func (a *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	//const op = "authService.Register"
	_, err := a.store.AddUser(&dto.AddUser{
		Login:    req.Login,
		Password: secret.GetHash(req.Password),
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{Message: "success"}, nil
}
func (a *AuthService) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	//const op = "authService.Unregister"
	tokenClaims, err := jwttoken.GetClaims(req.Token, a.secretKey)
	if err != nil {
		return nil, err
	}
	err = a.store.RemoveUser(tokenClaims.Sub)
	if err != nil {
		return nil, err
	}
	return &pb.UnregisterResponse{Message: "success"}, nil

}
func (a *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	//const op = "authService.Login"
	user, err := a.store.GetUserByLogin(req.Login)
	if err != nil {
		return nil, err
	}
	if !secret.CheckHash(req.Password, user.Password) {
		err := servererrors.ErrorInvalidLoginOrPassword
		return nil, err
	}

	tokenString, tokenClaims, err := jwttoken.CreateToken(user.UserId, req.DeviceCode, a.accessLifetime, a.secretKey)
	if err != nil {
		return nil, err
	}
	if err := a.store.AddTokenWithId(&dto.AddTokenWithId{
		TokenId:       tokenClaims.Jti,
		UserId:        user.UserId,
		DeviceCode:    req.DeviceCode,
		Token:         tokenString,
		TokenTypeCode: 'a',
		ExpirationAt:  tokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: tokenString}, nil
}
func (a *AuthService) CheckToken(ctx context.Context, req *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	//const op = "authService.CheckToken"
	_, err := jwttoken.GetToken(req.Token, a.secretKey)
	if err != nil {
		return nil, err
	}
	return &pb.CheckTokenResponse{Message: "success"}, nil

}
func (a *AuthService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	//const op = "authService.ChangePassword"
	tokenClaims, err := jwttoken.GetClaims(req.Token, a.secretKey)
	if err != nil {
		return nil, err
	}
	hashNewPassword := secret.GetHash(req.NewPassword)

	err = a.store.UpdateUser(&dto.UpdateUser{
		UserId:   tokenClaims.Sub,
		Password: &hashNewPassword,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ChangePasswordResponse{
		Message: "success",
	}, nil

}
