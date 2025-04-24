package servererrors

import "errors"

var (
	ErrorRecordNotFound         = errors.New("record not found")
	ErrorInternalServerError    = errors.New("internal server error")
	ErrorInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrorLoginAlreadyExists     = errors.New("login already exists")

	ErrorInvalidTokenSignature = errors.New("token signature is invalid")
	ErrorTokenExpired          = errors.New("token is either expired or not active yet")
	ErrorTokenMalformed        = errors.New("token malformed")
)
