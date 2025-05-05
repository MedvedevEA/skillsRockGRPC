package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddUser struct {
	Login    string
	Password string
}
type UpdateUser struct {
	UserId   *uuid.UUID
	Login    *string
	Password *string
}

type AddRefreshTokenWithRefreshTokenId struct {
	RefreshTokenId *uuid.UUID
	UserId         *uuid.UUID
	DeviceCode     string
	ExpirationAt   time.Time
	IsRevoke       bool
}
type RevokeRefreshTokensByUserIdAndDeviceCode struct {
	UserId     *uuid.UUID
	DeviceCode *string
}
