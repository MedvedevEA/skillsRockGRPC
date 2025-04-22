package dto

type Register struct {
	Login    string
	Password string
	Email    string
}

type Login struct {
	Login    string
	Password string
}

type CheckToken struct {
	Token string
}
