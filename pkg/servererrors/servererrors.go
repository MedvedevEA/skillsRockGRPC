package servererrors

import (
	"errors"
)

type SrvError struct {
	message string
	op      string
	err     error
}

func New(message string, op string, err error) SrvError {
	return SrvError{
		message,
		op,
		err,
	}
}
func (s *SrvError) Error() string {
	return s.message
}

var (
	ErrorRecordNotFound      = errors.New("record not found")
	ErrorInternalServerError = errors.New("internal server error")

	ErrorInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrorLoginAlreadyExists     = errors.New("login already exists")

	ErrorInvalidTokenSignature = errors.New("token signature is invalid")
	ErrorTokenExpired          = errors.New("token is either expired or not active yet")
	ErrorTokenMalformed        = errors.New("token malformed")
)
