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
	RemoveUserByLogin(login string) error

	GetRoleIdsByUserId(userId *uuid.UUID) ([]*uuid.UUID, error)

	AddTokenWithId(dto *dto.AddTokenWithId) error
	GetToken(tokenId *uuid.UUID) (*entity.Token, error)
	UpdateTokenRevokeByTokenId(tokenId *uuid.UUID) error
	UpdateTokensRevokeByUserId(userId *uuid.UUID) error
	RemoveTokensByUserIdAndDeviceCode(dto *dto.RemoveTokensByUserIdAndDeviceCode) error
}
