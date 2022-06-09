package token_store_contract

import (
	"errors"

	"github.com/k0marov/golang-auth/internal/data/models"
)

type TokenStore interface {
	FindUserFromToken(token string) (models.UserModel, error)
}

var TokenNotFoundErr = errors.New("token not found")
