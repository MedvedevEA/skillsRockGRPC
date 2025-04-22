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

func New(store repository.Repository, lg *slog.Logger, cfg *config.Token) *AuthService {
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
	_, err := a.store.AddUser(&dto.AddUser{
		Login:    req.Login,
		Password: secret.GetHash(req.Password),
		Email:    req.Email,
	})
	if err != nil {
		return &pb.RegisterResponse{Message: "failure"}, err
	}
	return &pb.RegisterResponse{Message: "success"}, nil
}
func (a *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := a.store.GetUserByLogin(req.Login)
	if err != nil {
		return nil, err
	}
	if !secret.CheckHash(req.Password, user.Password) {
		return nil, servererrors.ErrorInvalidUsernameOrPassword
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
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.CheckTokenResponse{Ok: token.Valid}, nil
}

/*




func (s *Service) Register(dto *srvDto.Register) (string, error) {
	hashPassword := secret.GetHash(dto.Password)
	_, err := s.store.AddUser(&repoDto.AddUser{
		Login:    dto.Login,
		Password: hashPassword,
		Email:    dto.Email,
	})
	if err != nil {
		return "", err
	}
	return "User registered successfully", nil
}
func (s *Service) Login(dto *srvDto.Login) (string, error) {
	user, err := s.store.GetUser(dto.Login)
	if err != nil {
		return "", err
	}
	if !secret.CheckHash(dto.Password, user.Password) {
		return "", servererrors.ErrorInvalidUsernameOrPassword
	}
	token, err := jwt.GenerateToken(user.Login, (*uuid.UUID)(s.secretKey))
	if err != nil {
		return "", servererrors.ErrorInternalServerError
	}
	return token, nil
}
func (s *Service) CheckToken(token string) (bool, error) {
	token, err := jwt.ParseToken(token, s.secretKey)
	if err !=nil
	s.lg.Info("", slog.Any("error", err))
	return err == nil, nil
}
*/
