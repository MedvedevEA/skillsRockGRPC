package authservice

import "errors"

var (
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrLoginAlreadyExists     = errors.New("login already exists")
	ErrInvalidArgumentUserId  = errors.New("invalid user id value")
	ErrInvalidArgumentTokenId = errors.New("invalid token id value")

	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found")
)
