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
	keyRSA, err := secure.UnmarshalRSAPrivate(keyByteArray)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	key := secure.MarshalRSAPrivate(keyRSA, "")

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
		Password: secure.GetHash(req.Password),
		Email:    req.Email,
	}); err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			return nil, status.Error(codes.AlreadyExists, servererrors.ErrLoginAlreadyExists.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterResponse{}, nil
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
	//Для таблиц user и token установлено правило каскадного удаления записей - при удалении пользователя, удаляются все его токены

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
	roleIds, err := a.store.GetUserRolesByUserId(user.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.RemoveTokensByUserIdAndDeviceCode(&dto.RemoveTokensByUserIdAndDeviceCode{
		UserId:     user.UserId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, accessTokenClaims, err := jwt.CreateToken(user.UserId, req.DeviceCode, roleIds, a.accessLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddTokenWithTokenId(&dto.AddTokenWithTokenId{
		TokenId:       accessTokenClaims.Jti,
		UserId:        accessTokenClaims.Sub,
		DeviceCode:    accessTokenClaims.DeviceCode,
		Token:         accessTokenString,
		TokenTypeCode: 'a',
		ExpirationAt:  accessTokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(user.UserId, req.DeviceCode, roleIds, a.refrashLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddTokenWithTokenId(&dto.AddTokenWithTokenId{
		TokenId:       refreshTokenClaims.Jti,
		UserId:        refreshTokenClaims.Sub,
		DeviceCode:    refreshTokenClaims.DeviceCode,
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
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentUserId, op).Error())
	}
	if err := a.store.RemoveTokensByUserIdAndDeviceCode(&dto.RemoveTokensByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: &req.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LogoutResponse{}, nil
}
func (a *AuthService) TokenIsRevoke(ctx context.Context, req *pb.TokenIsRevokeRequest) (*pb.TokenIsRevokeResponse, error) {
	const op = "authService.TokenIsRevoke"
	tokenId, err := uuid.Parse(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrapf(servererrors.ErrInvalidArgumentTokenId, op).Error())
	}
	token, err := a.store.GetToken(&tokenId)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrTokenNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TokenIsRevokeResponse{IsRevoke: token.IsRevoke}, nil
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
	if err := a.store.UpdateTokensRevokeByUserIdAndDeviceCode(&dto.UpdateTokensRevokeByUserIdAndDeviceCode{
		UserId:     &userId,
		DeviceCode: nil,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdatePasswordResponse{}, nil
}
func (a *AuthService) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	const op = "authService.RevokeToken"
	tokenId, err := uuid.Parse(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentTokenId, op).Error())
	}
	if err := a.store.UpdateTokenRevokeByTokenId(&tokenId); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrTokenNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RevokeTokenResponse{}, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	const op = "authService.RefreshToken"
	tokenId, err := uuid.Parse(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(servererrors.ErrInvalidArgumentTokenId, op).Error())
	}
	token, err := a.store.GetToken(&tokenId)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, servererrors.ErrTokenNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if token.IsRevoke {
		return nil, status.Error(codes.Unauthenticated, servererrors.ErrTokenRevoked.Error())
	}
	if token.TokenTypeCode != 'r' {
		return nil, status.Error(codes.InvalidArgument, servererrors.ErrInvalidTokenType.Error())
	}
	roleIds, err := a.store.GetUserRolesByUserId(token.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.RemoveTokensByUserIdAndDeviceCode(&dto.RemoveTokensByUserIdAndDeviceCode{
		UserId:     token.UserId,
		DeviceCode: &token.DeviceCode,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//access token
	accessTokenString, accessTokenClaims, err := jwt.CreateToken(token.UserId, token.DeviceCode, roleIds, a.accessLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddTokenWithTokenId(&dto.AddTokenWithTokenId{
		TokenId:       accessTokenClaims.Jti,
		UserId:        accessTokenClaims.Sub,
		DeviceCode:    accessTokenClaims.DeviceCode,
		Token:         accessTokenString,
		TokenTypeCode: 'a',
		ExpirationAt:  accessTokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//refresh token
	refreshTokenString, refreshTokenClaims, err := jwt.CreateToken(token.UserId, token.DeviceCode, roleIds, a.refrashLifetime, a.key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := a.store.AddTokenWithTokenId(&dto.AddTokenWithTokenId{
		TokenId:       refreshTokenClaims.Jti,
		UserId:        refreshTokenClaims.Sub,
		DeviceCode:    refreshTokenClaims.DeviceCode,
		Token:         refreshTokenString,
		TokenTypeCode: 'r',
		ExpirationAt:  refreshTokenClaims.ExpiresAt.Time,
		IsRevoke:      false,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
