package repository

import (
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"

	"github.com/google/uuid"
)

type Repository interface {
	AddUser(dto *dto.AddUser) (*entity.User, error)
	GetUser(userId *uuid.UUID) (*entity.User, error)
	UpdateUser(dto *dto.UpdateUser) error
}
