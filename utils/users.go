package utils

import (
	"djanGO/db"
)

var (
	UserAuthenticateFunc = authenticateUser
	UserRegisterFunc     = registerUser
)

func RegisterUser(login, password string) error {
	return UserRegisterFunc(login, password)
}

func AuthenticateUser(login, password string) (bool, error) {
	return UserAuthenticateFunc(login, password)
}

func authenticateUser(login, password string) (bool, error) {
	return db.AuthenticateUser(login, password)
}

func registerUser(login, password string) error {
	return db.CreateUser(login, password)
}
