package auth

import (
	"fmt"

	"internal/core/crypto/bcrypt_hasher"
	"internal/data/store"
	"internal/data/store/db_file_interactor_impl"
	"internal/delivery/server"
	"internal/delivery/token_auth_middleware"
	"internal/domain/auth_service"
	"internal/domain/entities"
)

var UserContextKey = token_auth_middleware.UserContextKey{}

func NewStoreImpl(dbFileName string) (*store.PersistentInMemoryFileStore, error) {
	fileInteractor := db_file_interactor_impl.NewDBFileInteractor(dbFileName)

	store, err := store.NewPersistentInMemoryFileStore(fileInteractor)
	if err != nil {
		return nil, fmt.Errorf("problem creating a store: %v", err)
	}
	return store, nil
}

func NewAuthServerImpl(store *store.PersistentInMemoryFileStore, hashCost int, onNewRegister func(User)) (*server.AuthServer, error) {
	hasher := bcrypt_hasher.NewBcryptHasher(hashCost)
	service := auth_service.NewAuthServiceImpl(store, hasher, onNewRegister)
	return server.NewAuthServer(service), nil
}

func NewTokenAuthMiddleware(store *store.PersistentInMemoryFileStore) *token_auth_middleware.TokenAuthMiddleware {
	return token_auth_middleware.NewTokenAuthMiddleware(store)
}

type User = entities.User
