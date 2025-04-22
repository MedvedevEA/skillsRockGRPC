package repository

import (
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"

	"github.com/google/uuid"
)

type Repository interface {
	AddUser(dto *dto.AddUser) (*uuid.UUID, error)
	GetUser(login string) (*entity.User, error)
	UpdateUser(dto *dto.UpdateUser) error
}
