package authservice

import (
	"context"
	"crypto/rsa"
	"log"
	"log/slog"
	pb "skillsRockGRPC/grpc/genproto"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/repository"
	"skillsRockGRPC/internal/repository/dto"

	"skillsRockGRPC/pkg/jwt"
	"skillsRockGRPC/pkg/secure"
	"skillsRockGRPC/pkg/servererrors"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	store           repository.Repository
	privateKey      *rsa.PrivateKey
	accessLifetime  time.Duration
	refrashLifetime time.Duration
	lg              *slog.Logger
}

func MustNew(store repository.Repository, lg *slog.Logger, cfg *config.Token) *AuthService {
	const op = "authservice.MustNew"
	privateKey, err := secure.LoadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatalf("%s: %v", op, err)
	}

	return &AuthService{
		store:           store,
		privateKey:      privateKey,
		accessLifetime:  cfg.AccessLifetime,
		refrashLifetime: cfg.RefreshLifetime,
		lg:              lg,
	}
}

func (a *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	//const op = "authService.Register"
	userId, err := a.store.AddUser(&dto.AddUser{
		Login:    req.Login,
		Password: secure.GetHash(req.Password),
	})
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			return nil, status.Error(codes.AlreadyExists, servererrors.ErrLoginAlreadyExists.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterResponse{UserId: userId.String()}, nil
}
func (a *AuthService) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	const op = "authService.Unregister"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	if err := a.store.RemoveUser(&userId); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnregisterResponse{}, nil

}
func (a *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	const op = "authService.Login"
	if req.DeviceCode == "" {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentDeviceCode, op).Error())
	}
	user, err := a.store.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.Unauthenticated, servererrors.ErrInvalidLoginOrPassword.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !secure.CheckHash(req.Password, user.Password) {
		err := status.Error(codes.Unauthenticated, servererrors.ErrInvalidLoginOrPassword.Error())
		return nil, err
	}
	if err := a.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     user.UserId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, _, err := jwt.CreateToken(user.UserId, req.DeviceCode, "access", a.accessLifetime, a.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(user.UserId, req.DeviceCode, "refresh", a.refrashLifetime, a.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddRefreshTokenWithRefreshTokenId(&dto.AddRefreshTokenWithRefreshTokenId{
		RefreshTokenId: refreshTokenClaims.Jti,
		UserId:         refreshTokenClaims.Sub,
		DeviceCode:     refreshTokenClaims.DeviceCode,
		ExpirationAt:   refreshTokenClaims.ExpiresAt.Time,
		IsRevoke:       false,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}
func (a *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	const op = "authService.Logout"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	if req.DeviceCode == "" {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentDeviceCode, op).Error())
	}
	if err := a.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LogoutResponse{}, nil
}
func (a *AuthService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	const op = "authService.UpdatePassword"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	hashNewPassword := secure.GetHash(req.NewPassword)

	if err = a.store.UpdateUser(&dto.UpdateUser{
		UserId:   &userId,
		Password: &hashNewPassword,
	}); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: nil,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdatePasswordResponse{}, nil
}
func (a *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	const op = "authService.RefreshToken"
	refreshTokenId, err := uuid.Parse(req.RefreshTokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentTokenId, op).Error())
	}
	refreshToken, err := a.store.GetRefreshToken(&refreshTokenId)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrTokenNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if refreshToken.IsRevoke {
		if err := a.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
			UserId:     refreshToken.UserId,
			DeviceCode: &refreshToken.DeviceCode,
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Unauthenticated, servererrors.ErrTokenRevoked.Error())
	}
	if err := a.store.RevokeRefreshTokenByRefreshTokenId(&refreshTokenId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, _, err := jwt.CreateToken(refreshToken.UserId, refreshToken.DeviceCode, "access", a.accessLifetime, a.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(refreshToken.UserId, refreshToken.DeviceCode, "refresh", a.refrashLifetime, a.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddRefreshTokenWithRefreshTokenId(&dto.AddRefreshTokenWithRefreshTokenId{
		RefreshTokenId: refreshTokenClaims.Jti,
		UserId:         refreshTokenClaims.Sub,
		DeviceCode:     refreshTokenClaims.DeviceCode,
		ExpirationAt:   refreshTokenClaims.ExpiresAt.Time,
		IsRevoke:       false,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RefreshTokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
