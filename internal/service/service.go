package service

import (
	"crypto/rsa"
	"log"
	"log/slog"
	"os"
	"skillsRockGRPC/internal/repository"
	repoDto "skillsRockGRPC/internal/repository/dto"
	srvDto "skillsRockGRPC/internal/service/dto"
	"skillsRockGRPC/pkg/jwt"
	"skillsRockGRPC/pkg/secret"
	"skillsRockGRPC/pkg/servererrors"

	"github.com/pkg/errors"
)

type Service struct {
	store  repository.Repository
	secret *rsa.PrivateKey
	lg     *slog.Logger
}

func MustNew(store repository.Repository, lg *slog.Logger, secretPath string) *Service {
	const op = "service.MustNew"
	secretByteArray, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	secretRSA, err := secret.UnmarshalRSAPrivate(secretByteArray)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}

	return &Service{
		store:  store,
		secret: secretRSA,
		lg:     lg,
	}
}

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
		return "", servererrors.ErrInvalidUsernameOrPassword
	}
	token, err := jwt.GenerateToken(user.Login, secret.MarshalRSAPrivate(s.secret, ""))
	if err != nil {
		return "", servererrors.ErrInternalServerError
	}
	return token, nil
}
func (s *Service) CheckToken(dto *srvDto.CheckToken) (bool, error) {
	_, err := jwt.ParseToken(dto.Token, secret.MarshalRSAPrivate(s.secret, ""))
	s.lg.Info("", slog.Any("error", err))
	return err == nil, nil
}
