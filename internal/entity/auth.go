package entity

import "github.com/google/uuid"

type User struct {
	UserId   *uuid.UUID `json:"user_id"`
	Login    string     `json:"login"`
	Password string     `json:"password"`
	Email    string     `json:"email"`
}

type Service struct {
	ServiceID  *uuid.UUID `json:"service_id"`
	Name       string     `json:"name"`
	PrivateKey string     `json:"private_key"`
	PublicKey  string     `json:"public_key"`
}
