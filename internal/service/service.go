package service

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

type Service struct {
	pb.UnimplementedAuthServiceServer
	store           repository.Repository
	privateKey      *rsa.PrivateKey
	accessLifetime  time.Duration
	refrashLifetime time.Duration
	lg              *slog.Logger
}

func MustNew(store repository.Repository, lg *slog.Logger, cfg *config.Token) *Service {
	const op = "service.MustNew"
	privateKey, err := secure.LoadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatalf("%s: %v", op, err)
	}

	return &Service{
		store:           store,
		privateKey:      privateKey,
		accessLifetime:  cfg.AccessLifetime,
		refrashLifetime: cfg.RefreshLifetime,
		lg:              lg,
	}
}

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	//const op = "service.Register"
	userId, err := s.store.AddUser(&dto.AddUser{
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
func (s *Service) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	const op = "service.Unregister"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	if err := s.store.RemoveUser(&userId); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnregisterResponse{}, nil

}
func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	const op = "service.Login"
	if req.DeviceCode == "" {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentDeviceCode, op).Error())
	}
	user, err := s.store.GetUserByLogin(req.Login)
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
	if err := s.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     user.UserId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, _, err := jwt.CreateToken(user.UserId, req.DeviceCode, "access", s.accessLifetime, s.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(user.UserId, req.DeviceCode, "refresh", s.refrashLifetime, s.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.store.AddRefreshTokenWithRefreshTokenId(&dto.AddRefreshTokenWithRefreshTokenId{
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
func (s *Service) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	const op = "service.Logout"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	if req.DeviceCode == "" {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentDeviceCode, op).Error())
	}
	if err := s.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LogoutResponse{}, nil
}
func (s *Service) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	const op = "Service.UpdatePassword"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	hashNewPassword := secure.GetHash(req.NewPassword)

	if err = s.store.UpdateUser(&dto.UpdateUser{
		UserId:   &userId,
		Password: &hashNewPassword,
	}); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: nil,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdatePasswordResponse{}, nil
}
func (s *Service) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	const op = "service.RefreshToken"
	refreshTokenId, err := uuid.Parse(req.RefreshTokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentTokenId, op).Error())
	}
	refreshToken, err := s.store.GetRefreshToken(&refreshTokenId)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrTokenNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if refreshToken.IsRevoke {
		if err := s.store.RevokeRefreshTokensByUserIdAndDeviceCode(&dto.RevokeRefreshTokensByUserIdAndDeviceCode{
			UserId:     refreshToken.UserId,
			DeviceCode: &refreshToken.DeviceCode,
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Unauthenticated, servererrors.ErrTokenRevoked.Error())
	}
	if err := s.store.RevokeRefreshTokenByRefreshTokenId(&refreshTokenId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, _, err := jwt.CreateToken(refreshToken.UserId, refreshToken.DeviceCode, "access", s.accessLifetime, s.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(refreshToken.UserId, refreshToken.DeviceCode, "refresh", s.refrashLifetime, s.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.store.AddRefreshTokenWithRefreshTokenId(&dto.AddRefreshTokenWithRefreshTokenId{
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
