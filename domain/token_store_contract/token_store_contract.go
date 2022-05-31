package token_store_contract

import (
	"auth/domain/entities"
	"errors"
)

type TokenStore interface {
	FindUserFromToken(token string) (entities.StoredUser, error)
}

var TokenNotFoundErr = errors.New("Token not found")
