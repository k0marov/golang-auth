package auth_service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/domain/auth_store_contract"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/values"

	"github.com/google/uuid"
)

type AuthStore = auth_store_contract.AuthStore

type Hasher interface {
	IsHashed(password string) bool
	Hash(password string) (string, error)
	Compare(pass, hashedPass string) bool
}

type AuthServiceImpl struct {
	store  AuthStore
	hasher Hasher
}

func NewAuthServiceImpl(store AuthStore, hasher Hasher) *AuthServiceImpl {
	return &AuthServiceImpl{
		store:  store,
		hasher: hasher,
	}
}

func (s *AuthServiceImpl) Register(authData values.AuthData) (entities.Token, error) {
	if !checkUsernameValidity(authData.Username) {
		return entities.Token{}, client_errors.UsernameInvalidError
	}
	if s.store.UserExists(authData.Username) {
		return entities.Token{}, client_errors.UsernameAlreadyTakenError
	}
	if !s.hasher.IsHashed(authData.Password) {
		return entities.Token{}, client_errors.UnhashedPasswordError
	}

	hashedPassword, err := s.hasher.Hash(authData.Password)
	if err != nil {
		return entities.Token{}, errors.New(fmt.Sprintf("Error while hashing password: %v", err))
	}
	token := generateToken()
	_, err = s.store.CreateUser(authData.Username, string(hashedPassword), token)
	if err != nil {
		return entities.Token{}, errors.New(fmt.Sprintf("Error while creating a new user: %v", err))
	}
	return token, nil
}

func (s *AuthServiceImpl) Login(authData values.AuthData) (entities.Token, error) {
	existingUser, err := s.store.FindUser(authData.Username)
	if err != nil {
		if err == auth_store_contract.UserNotFoundErr {
			return entities.Token{}, client_errors.InvalidCredentialsError
		} else {
			return entities.Token{}, err
		}
	}
	if !s.hasher.Compare(authData.Password, existingUser.StoredPass) {
		return entities.Token{}, client_errors.InvalidCredentialsError
	}

	return existingUser.AuthToken, nil
}

const ValidUsernameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"
const MaxUsernameLength = 20

func checkUsernameValidity(username string) bool {
	if len(username) > MaxUsernameLength || username == "" {
		return false
	}
	if username[0] == byte('_') {
		return false
	}
	for _, char := range username {
		if !strings.Contains(ValidUsernameChars, string(char)) {
			return false
		}
	}
	return true
}

func generateToken() entities.Token {
	// this actually never returns an error
	token, _ := uuid.NewUUID()
	return entities.Token{Token: token.String()}
}
