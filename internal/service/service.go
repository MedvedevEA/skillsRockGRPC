package service

import (
	"crypto/rsa"
	"log"
	"log/slog"
	"os"
	"skillsRockGRPC/internal/repository"
	"skillsRockGRPC/pkg/secret"

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

func (s *Service) Register(dto *Service) (string, error) {
	return "", nil
}
func (s *Service) Login(dto *Service) (string, error) {
	return "", nil
}
func (s *Service) CheckToken(dto *Service) (bool, error) {
	return false, nil
}
