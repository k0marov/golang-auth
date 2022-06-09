package token_auth_middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/data/models"
	"github.com/k0marov/golang-auth/internal/delivery/token_auth_middleware"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/domain/mappers"
	"github.com/k0marov/golang-auth/internal/domain/token_store_contract"
	. "github.com/k0marov/golang-auth/internal/test_helpers"
)

func TestTokenAuthMiddleware(t *testing.T) {
	var validToken = "abracadabra"
	var storedUserWithThisToken = models.UserModel{
		Id:         RandomInt(),
		Username:   "John",
		StoredPass: RandomString(),
		AuthToken:  entities.Token{Token: validToken},
	}
	var userWithThisToken = mappers.ModelToUser(storedUserWithThisToken)

	store := &StubTokenStore{
		findUserFromToken: func(token string) (models.UserModel, error) {
			if token == validToken {
				return storedUserWithThisToken, nil
			} else {
				return models.UserModel{}, token_store_contract.TokenNotFoundErr
			}
		},
	}
	t.Run("should set requests's Context {User} key", func(t *testing.T) {
		createMiddleware := func(spyHandler *SpyHTTPHandler, store *StubTokenStore) http.Handler {
			return token_auth_middleware.NewTokenAuthMiddleware(store).Middleware(spyHandler)
		}
		t.Run("happy case (valid token is provided)", func(t *testing.T) {
			spyHandler := &SpyHTTPHandler{}
			middleware := createMiddleware(spyHandler, store)

			request := httptest.NewRequest(http.MethodGet, "/some/random/url", nil)
			request.Header.Set("Authorization", "Token "+validToken)
			response := httptest.NewRecorder()
			middleware.ServeHTTP(response, request)

			assertCalls(t, spyHandler, 1)
			updatedRequest := spyHandler.calls[0].r
			userInContext := updatedRequest.Context().Value(token_auth_middleware.UserKey)
			Assert(t, userInContext.(entities.User), userWithThisToken, "user in context")
		})
		t.Run("error case (some database error happened)", func(t *testing.T) {
			spyHandler := &SpyHTTPHandler{}
			errorStore := &StubTokenStore{
				findUserFromToken: func(string) (models.UserModel, error) {
					return models.UserModel{}, errors.New(RandomString())
				},
			}
			middleware := createMiddleware(spyHandler, errorStore)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/some/random/url", nil)
			request.Header.Set("Authorization", "Token "+validToken)
			middleware.ServeHTTP(response, request)

			assertCalls(t, spyHandler, 0)
			Assert(t, response.Result().StatusCode, http.StatusInternalServerError, "response status code")
		})
		t.Run("error case (no token is provided)", func(t *testing.T) {
			spyHandler := &SpyHTTPHandler{}
			middleware := createMiddleware(spyHandler, store)

			response := httptest.NewRecorder()
			middleware.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/some/random/url", nil))

			assertCalls(t, spyHandler, 0)
			AssertHTTPError(t, response, client_errors.AuthTokenRequiredError, http.StatusUnauthorized)
		})
		t.Run("error case (provided token is invalid)", func(t *testing.T) {
			spyHandler := &SpyHTTPHandler{}
			middleware := createMiddleware(spyHandler, store)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/some/random/url", nil)
			request.Header.Set("Authorization", "Token INVALIDTOKEN")
			middleware.ServeHTTP(response, request)

			assertCalls(t, spyHandler, 0)
			AssertHTTPError(t, response, client_errors.AuthTokenInvalidError, http.StatusUnauthorized)
		})
	})
}

func assertCalls(t testing.TB, spyHandler *SpyHTTPHandler, amountOfCalls int) {
	t.Helper()
	AssertFatal(t, len(spyHandler.calls), amountOfCalls, "amount of calls to next handler")
}

type handleCall struct {
	w http.ResponseWriter
	r *http.Request
}

type SpyHTTPHandler struct {
	calls []handleCall
}

func (s *SpyHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.calls = append(s.calls, handleCall{w, r})
}

type StubTokenStore struct {
	findUserFromToken func(string) (models.UserModel, error)
}

func (s *StubTokenStore) FindUserFromToken(token string) (models.UserModel, error) {
	if s.findUserFromToken != nil {
		return s.findUserFromToken(token)
	} else {
		return models.UserModel{}, token_store_contract.TokenNotFoundErr
	}
}
