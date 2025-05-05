package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserId   *uuid.UUID `json:"user_id" db:"user_id"`
	Login    string     `json:"login" db:"login"`
	Password string     `json:"password" db:"password"`
}

type RefreshToken struct {
	RefreshTokenId *uuid.UUID `json:"refresh_token_id" db:"refresh_token_id"`
	UserId         *uuid.UUID `json:"user_id" db:"user_id"`
	DeviceCode     string     `json:"device_code" db:"device_code"`
	ExpirationAt   time.Time  `json:"expiration_at" db:"expiration_at"`
	IsRevoke       bool       `json:"is_revoke" db:"is_revoke"`
}
