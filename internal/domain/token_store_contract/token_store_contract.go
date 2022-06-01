package token_store_contract

import (
	"errors"

	"github.com/k0marov/golang-auth/internal/domain/entities"
)

type TokenStore interface {
	FindUserFromToken(token string) (entities.StoredUser, error)
}

var TokenNotFoundErr = errors.New("Token not found")
