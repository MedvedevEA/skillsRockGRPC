package servererrors

import "errors"

var (
	ErrRecordNotFound            = errors.New("record not found")
	ErrInternalServerError       = errors.New("internal server error")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)
