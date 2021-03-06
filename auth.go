package auth

import (
	"fmt"
	"net/http"

	"github.com/k0marov/golang-auth/internal/core/crypto/bcrypt_hasher"
	"github.com/k0marov/golang-auth/internal/data/store"
	"github.com/k0marov/golang-auth/internal/data/store/db_file_interactor_impl"
	"github.com/k0marov/golang-auth/internal/delivery/http/handlers"
	"github.com/k0marov/golang-auth/internal/delivery/token_auth_middleware"
	"github.com/k0marov/golang-auth/internal/domain/auth_service"
	"github.com/k0marov/golang-auth/internal/domain/entities"
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

func NewHandlersImpl(store *store.PersistentInMemoryFileStore, hashCost int, onNewRegister func(User)) (login http.Handler, register http.Handler) {
	hasher := bcrypt_hasher.NewBcryptHasher(hashCost)
	service := auth_service.NewAuthServiceImpl(store, hasher, onNewRegister)
	return handlers.NewLoginHandler(service.Login), handlers.NewRegisterHandler(service.Register)
}

func NewTokenAuthMiddleware(store *store.PersistentInMemoryFileStore) *token_auth_middleware.TokenAuthMiddleware {
	return token_auth_middleware.NewTokenAuthMiddleware(store)
}

type User = entities.User
