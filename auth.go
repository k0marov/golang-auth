package auth

import (
	"fmt"
	"os"

	"github.com/k0marov/golang-auth/internal/core/crypto/bcrypt_hasher"
	"github.com/k0marov/golang-auth/internal/data/store"
	"github.com/k0marov/golang-auth/internal/data/store/db_file_interactor_impl"
	"github.com/k0marov/golang-auth/internal/delivery/server"
	"github.com/k0marov/golang-auth/internal/delivery/token_auth_middleware"
	"github.com/k0marov/golang-auth/internal/domain/auth_service"
	"github.com/k0marov/golang-auth/internal/domain/entities"
)

var UserContextKey = token_auth_middleware.UserContextKey{}

func NewStoreImpl(dbFileName string) (*store.PersistentInMemoryFileStore, error) {
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("problem opening %s %v", dbFileName, err)
	}
	fileInteractor := db_file_interactor_impl.NewDBFileInteractor(dbFile)

	store, err := store.NewPersistentInMemoryFileStore(fileInteractor)
	if err != nil {
		return nil, fmt.Errorf("problem creating a store: %v", err)
	}
	return store, nil
}

func NewAuthServerImpl(store *store.PersistentInMemoryFileStore, hashCost int) (*server.AuthServer, error) {
	hasher := bcrypt_hasher.NewBcryptHasher(hashCost)
	service := auth_service.NewAuthServiceImpl(store, hasher)
	return server.NewAuthServer(service), nil
}

func NewTokenAuthMiddleware(store *store.PersistentInMemoryFileStore) *token_auth_middleware.TokenAuthMiddleware {
	return token_auth_middleware.NewTokenAuthMiddleware(store)
}

type User = entities.User
