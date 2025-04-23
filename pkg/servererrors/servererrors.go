package servererrors

import "errors"

var (
	ErrorRecordNotFound         = errors.New("record not found")
	ErrorInternalServerError    = errors.New("internal server error")
	ErrorInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrorLoginAlreadyExists     = errors.New("login already exists")
	ErrorInvalidTokenSignature  = errors.New("token signature is invalid")
	ErrorInvalidTokenClaims     = errors.New("token claims is invalid")
)
