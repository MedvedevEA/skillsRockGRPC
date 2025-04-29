package repository

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")

	ErrRecordNotFound  = errors.New("record not found")
	ErrUniqueViolation = errors.New("unique violation")
)
