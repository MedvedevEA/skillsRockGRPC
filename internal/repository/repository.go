package repository

import (
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"

	"github.com/google/uuid"
)

type Repository interface {
	AddUser(dto *dto.AddUser) (*uuid.UUID, error)
	GetUser(userId *uuid.UUID) (*entity.User, error)
	GetUserByLogin(login string) (*entity.User, error)
	UpdateUser(dto *dto.UpdateUser) error
	RemoveUser(userId *uuid.UUID) error

	GetUserRolesByUserId(userId *uuid.UUID) ([]*uuid.UUID, error)

	AddTokenWithTokenId(dto *dto.AddTokenWithTokenId) error
	GetToken(tokenId *uuid.UUID) (*entity.Token, error)
	UpdateTokenRevokeByTokenId(tokenId *uuid.UUID) error
	UpdateTokensRevokeByUserIdAndDeviceCode(dto *dto.UpdateTokensRevokeByUserIdAndDeviceCode) error
	RemoveTokensByUserIdAndDeviceCode(dto *dto.RemoveTokensByUserIdAndDeviceCode) error
}

const (
	UpdateTokenRevokeByTokenIdQuery = `
		UPDATE "token" SET is_valid=false WHERE token_id=$1 RETURNING token_id;
	`
	UpdateTokenRevokeByUserIdAndDevceCodeQuery = `
		UPDATE "token" SET is_valid=false WHERE user_id=$1 AND ($2 IS NULL OR device_code=$2);
	`
	removeTokenByUserIdAndDeviceCodeQuery = `
		DELETE FROM "token" WHERE user_id=$1 AND ($2 IS NULL OR device_code=$2);
	`
)
