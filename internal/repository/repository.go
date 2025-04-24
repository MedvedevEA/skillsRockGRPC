package repository

import (
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"

	"github.com/google/uuid"
)

type Repository interface {
	AddUser(dto *dto.AddUser) (*uuid.UUID, error)
	GetUserByLogin(login string) (*entity.User, error)
	UpdateUser(dto *dto.UpdateUser) error
	RemoveUser(userId *uuid.UUID) error

	GetRolesByUserId(userId *uuid.UUID) ([]string, error)

	AddTokenWithId(dto *dto.AddTokenWithId) error
	GetToken(tokenId *uuid.UUID) (*entity.Token, error)
	UpdateTokenRevokeByUserIdAndDeviceCode(dto *dto.UpdateTokenRevokeByUserIdAndDeviceCode) error
	UpdateTokenRevokeByUserId(userId *uuid.UUID) error
	RemoveTokenByUserIdAndDeviceCode(dto *dto.RemoveTokenByUserIdAndDeviceCode) error
}
