package token_auth_middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/domain/token_store_contract"
)

func NewTokenAuthMiddleware(tokenStore token_store_contract.TokenStore) *TokenAuthMiddleware {
	return &TokenAuthMiddleware{tokenStore}
}

type TokenAuthMiddleware struct {
	tokenStore token_store_contract.TokenStore
}

func (t *TokenAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Token ")
		if authToken != "" {
			storedUser, err := t.tokenStore.FindUserFromToken(authToken)
			if err != nil {
				if err == token_store_contract.TokenNotFoundErr {
					throwUnauthorized(w, client_errors.AuthTokenInvalidError)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			user := entities.User{
				Id:       storedUser.Id,
				Username: storedUser.Username,
			}
			newContext := context.WithValue(r.Context(), "User", user)
			next.ServeHTTP(w, r.WithContext(newContext))
		} else {
			throwUnauthorized(w, client_errors.AuthTokenRequiredError)
		}
	})
}

func throwUnauthorized(w http.ResponseWriter, error client_errors.ClientError) {
	errorBuf := bytes.NewBuffer(nil)
	json.NewEncoder(errorBuf).Encode(error)
	http.Error(w, errorBuf.String(), http.StatusUnauthorized)
}
