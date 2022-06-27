package auth_service

import (
	"fmt"
	"strings"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/domain/auth_store_contract"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/domain/mappers"
	"github.com/k0marov/golang-auth/internal/values"

	"github.com/google/uuid"
)

type AuthStore = auth_store_contract.AuthStore

type Hasher interface {
	Hash(password string) (string, error)
	Compare(pass, hashedPass string) bool
}

type AuthServiceImpl struct {
	store         AuthStore
	hasher        Hasher
	onNewRegister func(entities.User)
}

// The onNewRegister function is called every time a new user is registered.
// This function can be used, for example, for creating a User Profile in some other database.
// It is called synchronously, which can be slow if it does something expensive.
// So, if you don't need synchronous behavior for this handler, wrap the expensive operation in a goroutine
func NewAuthServiceImpl(store AuthStore, hasher Hasher, onNewRegister func(entities.User)) *AuthServiceImpl {
	return &AuthServiceImpl{
		store:         store,
		hasher:        hasher,
		onNewRegister: onNewRegister,
	}
}

func (s *AuthServiceImpl) Register(authData values.AuthData) (entities.Token, error) {
	if !checkUsernameValidity(authData.Username) {
		return entities.Token{}, client_errors.UsernameInvalidError
	}
	if s.store.UserExists(authData.Username) {
		return entities.Token{}, client_errors.UsernameAlreadyTakenError
	}

	hashedPassword, err := s.hasher.Hash(authData.Password)
	if err != nil {
		return entities.Token{}, fmt.Errorf("error while hashing password: %w", err)
	}
	token := generateToken()
	newUser, err := s.store.CreateUser(authData.Username, string(hashedPassword), token)
	if err != nil {
		return entities.Token{}, fmt.Errorf("error while creating a new user: %w", err)
	}

	s.onNewRegister(mappers.ModelToUser(newUser))

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
