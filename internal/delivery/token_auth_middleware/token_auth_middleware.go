package token_auth_middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"internal/core/client_errors"
	"internal/domain/mappers"
	"internal/domain/token_store_contract"
)

func NewTokenAuthMiddleware(tokenStore token_store_contract.TokenStore) *TokenAuthMiddleware {
	return &TokenAuthMiddleware{tokenStore}
}

type UserContextKey struct{}

var UserKey = UserContextKey{}

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
			user := mappers.ModelToUser(storedUser)
			newContext := context.WithValue(r.Context(), UserContextKey{}, user)
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
