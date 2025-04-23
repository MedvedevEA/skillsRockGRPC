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

	"skillsRockGRPC/pkg/secret"
	"skillsRockGRPC/pkg/servererrors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
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

	tokenId := uuid.New()
	now := time.Now()
	expirionAt := now.Add(a.accessLifetime)
	claims := jwt.MapClaims{
		"deviceCode": req.DeviceCode,

		"sub": user.UserId,       // subject — субъект, которому выдан токен
		"exp": expirionAt.Unix(), // expiration time — время, когда токен станет невалидным
		"nbf": now.Unix(),        // not before — время, с которого токен должен считаться действительным
		"jat": now.Unix(),        // issued at — время, в которое был выдан токен
		"jti": tokenId,           // JWT ID — уникальный идентификатор токена
	}
	tokenJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tokenJwt.SignedString(a.secretKey)
	if err != nil {
		return nil, err
	}
	if err := a.store.AddTokenWithId(&dto.AddTokenWithId{
		TokenId:       &tokenId,
		UserId:        user.UserId,
		DeviceCode:    req.DeviceCode,
		Token:         tokenString,
		TokenTypeCode: 'a',
		ExpirationAt:  expirionAt,
		IsRevoke:      false,
	}); err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: tokenString}, nil
}
func (a *AuthService) CheckToken(ctx context.Context, req *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	//const op = "authService.CheckToken"
	_, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.CheckTokenResponse{Message: "success"}, nil

}
func (a *AuthService) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	//const op = "authService.Unregister"
	tokenJwt, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tokenJwt.Claims.(jwt.MapClaims)
	if !ok {
		return nil, servererrors.ErrorInvalidTokenClaims
	}
	userId, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, servererrors.ErrorInvalidTokenClaims
	}
	err = a.store.RemoveUser(&userId)
	if err != nil {
		return nil, err
	}
	return &pb.UnregisterResponse{Message: "success"}, nil

}
