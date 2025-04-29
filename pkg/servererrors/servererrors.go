package servererrors

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidLoginOrPassword    = errors.New("invalid login or password")
	ErrLoginAlreadyExists        = errors.New("login already exists")
	ErrInvalidArgumentUserId     = errors.New("invalid user id value")
	ErrInvalidArgumentTokenId    = errors.New("invalid token id value")
	ErrInvalidArgumentDeviceCode = errors.New("invalid device code value")
	ErrTokenRevoked              = errors.New("token is revoke")
	ErrInvalidTokenType          = errors.New("invalid token type")

	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found")
)
