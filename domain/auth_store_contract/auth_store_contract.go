package auth_store_contract

import (
	"auth/domain/entities"
	"errors"
)

type AuthStore interface {
	UserExists(username string) bool
	CreateUser(username string, storedPassword string, token entities.Token) (entities.StoredUser, error)
	FindUser(username string) (entities.StoredUser, error)
}

var UserNotFoundErr = errors.New("User not found")
