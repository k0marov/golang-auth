package auth_store_contract

import (
	"errors"

	"github.com/k0marov/golang-auth/internal/domain/entities"
)

type AuthStore interface {
	UserExists(username string) bool
	CreateUser(username string, storedPassword string, token entities.Token) (entities.StoredUser, error)
	FindUser(username string) (entities.StoredUser, error)
}

var UserNotFoundErr = errors.New("User not found")
