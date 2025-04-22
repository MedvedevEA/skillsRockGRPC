package dto

import (
	"github.com/google/uuid"
)

type AddUser struct {
	Login    string
	Password string
	Email    string
}
type UpdateUser struct {
	UserId   *uuid.UUID
	Login    *string
	Password *string
	Email    *string
}
