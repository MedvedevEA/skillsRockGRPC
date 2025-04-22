package servererrors

import "errors"

var (
	ErrorRecordNotFound            = errors.New("record not found")
	ErrorInternalServerError       = errors.New("internal server error")
	ErrorInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrorUsernameAlreadyExists     = errors.New("user name already exists")
)
