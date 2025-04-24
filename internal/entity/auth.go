package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserId   *uuid.UUID `json:"user_id" db:"user_id"`
	Login    string     `json:"login" db:"login"`
	Password string     `json:"password" db:"password"`
	Email    string     `json:"email" db:"email"`
}

type Role struct {
	RoleId *uuid.UUID `json:"role_id" db:"role_id"`
	Name   string     `json:"name" db:"name"`
}
type Token struct {
	TokenId       *uuid.UUID `json:"token_id" db:"token_id"`
	UserId        *uuid.UUID `json:"user_id" db:"user_id"`
	DeviceCode    *uuid.UUID `json:"device_code" db:"device_code"`
	Token         string     `json:"token" db:"token"`
	TokenTypeCode rune       `json:"token_type_code" db:"token_type_code"`
	ExpirationAt  time.Time  `json:"expiration_at" db:"expiration_at"`
	IsRevoke      bool       `json:"is_revoke" db:"is_revoke"`
}
type TokenType struct {
	TokenTypeCode rune   `json:"token_type_code" db:"token_type_code"`
	Name          string `json:"name" db:"name"`
}
