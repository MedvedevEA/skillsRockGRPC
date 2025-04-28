package authservice

import (
	"context"
	"log"
	"log/slog"
	"os"
	pb "skillsRockGRPC/gen/go/auth/v3"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/repository"
	"skillsRockGRPC/internal/repository/dto"

	"skillsRockGRPC/pkg/jwttoken"
	"skillsRockGRPC/pkg/secret"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	store           repository.Repository
	key             []byte
	accessLifetime  time.Duration
	refrashLifetime time.Duration
	lg              *slog.Logger
}

func MustNew(store repository.Repository, lg *slog.Logger, cfg *config.Token) *AuthService {
	const op = "authservice.MustNew"
	keyByteArray, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	keyRSA, err := secret.UnmarshalRSAPrivate(keyByteArray)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	key := secret.MarshalRSAPrivate(keyRSA, "")

	return &AuthService{
		store:           store,
		key:             key,
		accessLifetime:  cfg.AccessLifetime,
		refrashLifetime: cfg.RefreshLifetime,
		lg:              lg,
	}
}

func (a *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	//const op = "authService.Register"
	if _, err := a.store.AddUser(&dto.AddUser{
		Login:    req.Login,
		Password: secret.GetHash(req.Password),
		Email:    req.Email,
	}); err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			return nil, status.Error(codes.AlreadyExists, ErrLoginAlreadyExists.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterResponse{}, nil
}
func (a *AuthService) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	//const op = "authService.Unregister"
	if err := a.store.RemoveUserByLogin(req.Login); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	//Для таблиц user и token установлено правило каскадного удаления записей - при удалении пользователя, удаляются все его токены

	return &pb.UnregisterResponse{}, nil

}
func (a *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	//const op = "authService.Login"
	user, err := a.store.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !secret.CheckHash(req.Password, user.Password) {
		err := status.Error(codes.Unauthenticated, ErrInvalidLoginOrPassword.Error())
		return nil, err
	}
	roleIds, err := a.store.GetRoleIdsByUserId(user.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.RemoveTokensByUserIdAndDeviceCode(&dto.RemoveTokensByUserIdAndDeviceCode{
		UserId:     user.UserId,
		DeviceCode: req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, accessTokenClaims, err := jwttoken.CreateToken(user.UserId, req.DeviceCode, roleIds, a.accessLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwttoken.CreateToken(user.UserId, req.DeviceCode, roleIds, a.refrashLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}
func (a *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	const op = "authService.Logout"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(ErrInvalidArgumentUserId, op).Error())
	}
	if err := a.store.RemoveTokensByUserIdAndDeviceCode(&dto.RemoveTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LogoutResponse{}, nil
}
func (a *AuthService) TokenIsRevoke(ctx context.Context, req *pb.TokenIsRevokeRequest) (*pb.TokenIsRevokeResponse, error) {
	const op = "authService.TokenIsRevoke"
	tokenId, err := uuid.Parse(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrapf(ErrInvalidArgumentTokenId, op).Error())
	}
	token, err := a.store.GetToken(&tokenId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TokenIsRevokeResponse{IsRevoke: token.IsRevoke}, nil
}
func (a *AuthService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	const op = "authService.UpdatePassword"
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(ErrInvalidArgumentUserId, op).Error())
	}
	hashNewPassword := secret.GetHash(req.NewPassword)

	if err = a.store.UpdateUser(&dto.UpdateUser{
		UserId:   &userId,
		Password: &hashNewPassword,
	}); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, ErrInvalidArgumentUserId.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.UpdateTokensRevokeByUserId(&userId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdatePasswordResponse{}, nil
}
func (a *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	const op = "authService.RevokeToken"
	tokenId, err := uuid.Parse(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(ErrInvalidArgumentTokenId, op).Error())
	}
	if err := a.store.UpdateTokenRevokeByTokenId(&tokenId); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, ErrInvalidArgumentTokenId.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RevokeTokenResponse{}, nil
}
