package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserId   *uuid.UUID `json:"user_id"`
	Login    string     `json:"login"`
	Password string     `json:"password"`
	Email    string     `json:"email"`
}

type Role struct {
	RoleId *uuid.UUID `json:"role_id"`
	Name   string     `json:"name"`
}
type Token struct {
	TokenId       *uuid.UUID `json:"token_id"`
	UserId        *uuid.UUID `json:"user_id"`
	DeviceCode    *uuid.UUID `json:"device_code"`
	Token         string     `json:"token"`
	TokenTypeCode rune       `json:"token_type_code"`
	ExpirationAt  time.Time  `json:"expiration_at"`
	IsRevoke      bool       `json:"is_revoke"`
}
type TokenType struct {
	TokenTypeCode rune   `json:"token_type_code"`
	Name          string `json:"name"`
}
