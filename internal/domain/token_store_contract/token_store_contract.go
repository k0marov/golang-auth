package token_store_contract

import (
	"errors"

	"internal/data/models"
)

type TokenStore interface {
	FindUserFromToken(token string) (models.UserModel, error)
}

var TokenNotFoundErr = errors.New("token not found")
