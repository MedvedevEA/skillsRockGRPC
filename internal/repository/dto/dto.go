package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddUser struct {
	Login    string
	Password string
	Email    string
}
type UpdateUser struct {
	UserId   *uuid.UUID
	Login    *string
	Password *string
	Email    *string
}

type AddTokenWithId struct {
	TokenId       *uuid.UUID
	UserId        *uuid.UUID
	DeviceCode    string
	Token         string
	TokenTypeCode rune
	ExpirationAt  time.Time
	IsRevoke      bool
}
type RemoveTokensByUserIdAndDeviceCode struct {
	UserId     *uuid.UUID
	DeviceCode string
}
