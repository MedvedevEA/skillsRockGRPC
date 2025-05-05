package repository

import (
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	AddUser(dto *dto.AddUser) (*uuid.UUID, error)
	GetUserByLogin(login string) (*entity.User, error)
	UpdateUser(dto *dto.UpdateUser) error
	RemoveUser(userId *uuid.UUID) error

	AddRefreshTokenWithRefreshTokenId(dto *dto.AddRefreshTokenWithRefreshTokenId) error
	GetRefreshToken(refreshTokenId *uuid.UUID) (*entity.RefreshToken, error)
	RevokeRefreshTokenByRefreshTokenId(refreshTokenId *uuid.UUID) error
	RevokeRefreshTokensByUserIdAndDeviceCode(dto *dto.RevokeRefreshTokensByUserIdAndDeviceCode) error
	RemoveRefreshTokensByExpirationAt(now time.Time) (int64, error)
	//RemoveRefreshToken(refreshTokenId *uuid.UUID) (*entity.RefreshToken, error)
	//RemoveRefreshTokensByUserIdAndDeviceCode(dto *dto.RemoveRefreshTokensByUserIdAndDeviceCode) error
}
