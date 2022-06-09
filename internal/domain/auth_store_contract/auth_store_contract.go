package auth_store_contract

import (
	"errors"

	"internal/data/models"
	"internal/domain/entities"
)

type AuthStore interface {
	UserExists(username string) bool
	CreateUser(username string, storedPassword string, token entities.Token) (models.UserModel, error)
	FindUser(username string) (models.UserModel, error)
}

var UserNotFoundErr = errors.New("User not found")
