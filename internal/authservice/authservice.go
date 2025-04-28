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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, status.Error(codes.Internal, err.Error())
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
	//Для таблиц user и token установлено правило каскадного удаления записей.

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
	roleNames, err := a.store.GetRolesByUserId(user.UserId)
	if err != nil {
		return nil, err
	}
	err = a.store.RemoveTokenByUserIdAndDeviceCode(&dto.RemoveTokenByUserIdAndDeviceCode{
		UserId:     user.UserId,
		DeviceCode: req.DeviceCode,
	})
	if err != nil {
		return nil, err
	}
	//access token
	accessTokenString, accessTokenClaims, err := jwttoken.CreateToken(user.UserId, req.DeviceCode, roleNames, a.accessLifetime, a.secretKey)
	if err != nil {
		return nil, err
	}
	if err := a.store.AddTokenWithId(&dto.AddTokenWithId{
		TokenId:       accessTokenClaims.Jti,
		UserId:        user.UserId,
		DeviceCode:    req.DeviceCode,
		Token:         accessTokenString,
		TokenTypeCode: 'a',
		ExpirationAt:  accessTokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, err
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwttoken.CreateToken(user.UserId, req.DeviceCode, roleNames, a.refrashLifetime, a.secretKey)
	if err != nil {
		return nil, err
	}
	if err := a.store.AddTokenWithId(&dto.AddTokenWithId{
		TokenId:       refreshTokenClaims.Jti,
		UserId:        user.UserId,
		DeviceCode:    req.DeviceCode,
		Token:         refreshTokenString,
		TokenTypeCode: 'r',
		ExpirationAt:  refreshTokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, err
	}

	return &pb.LoginResponse{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}
func (a *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	//const op = "authService.Logout"
	tokenClaims, err := jwttoken.GetClaims(req.Token, a.secretKey)
	if err != nil {
		return nil, err
	}
	err = a.store.RemoveTokenByUserIdAndDeviceCode(&dto.RemoveTokenByUserIdAndDeviceCode{
		UserId:     tokenClaims.Sub,
		DeviceCode: tokenClaims.DeviceCode,
	})
	if err != nil {
		return nil, servererrors.ErrorInternalServerError
	}
	return &pb.LogoutResponse{Message: "success"}, nil
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
	err = a.store.UpdateTokenRevokeByUserId(tokenClaims.Sub)
	if err != nil {
		return nil, err
	}
	return &pb.ChangePasswordResponse{
		Message: "success",
	}, nil

}
func (a *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	//const op = "authService.RevokeToken"
	tokenClaims, err := jwttoken.GetClaims(req.Token, a.secretKey)
	if err != nil {
		return nil, err
	}
	err = a.store.UpdateTokenRevokeByUserIdAndDeviceCode(&dto.UpdateTokenRevokeByUserIdAndDeviceCode{
		UserId:     tokenClaims.Sub,
		DeviceCode: tokenClaims.DeviceCode,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RevokeTokenResponse{
		Message: "success",
	}, nil
}
