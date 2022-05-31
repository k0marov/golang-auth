package token_auth_middleware_test

import (
	"auth/core/client_errors"
	"auth/delivery/token_auth_middleware"
	"auth/domain/entities"
	"auth/domain/token_store_contract"
	. "auth/test_helpers"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenAuthMiddleware(t *testing.T) {
	var validToken = "abracadabra"
	var storedUserWithThisToken = entities.StoredUser{
		Id:         RandomString(),
		Username:   "John",
		StoredPass: RandomString(),
		AuthToken:  entities.Token{Token: validToken},
	}
	var userWithThisToken = entities.User{
		Id:       storedUserWithThisToken.Id,
		Username: storedUserWithThisToken.Username,
	}

	store := &StubTokenStore{
		findUserFromToken: func(token string) (entities.StoredUser, error) {
			if token == validToken {
				return storedUserWithThisToken, nil
			} else {
				return entities.StoredUser{}, token_store_contract.TokenNotFoundErr
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
			userInContext := updatedRequest.Context().Value("User")
			Assert(t, userInContext.(entities.User), userWithThisToken, "user in context")
		})
		t.Run("error case (some database error happened)", func(t *testing.T) {
			spyHandler := &SpyHTTPHandler{}
			errorStore := &StubTokenStore{
				findUserFromToken: func(string) (entities.StoredUser, error) {
					return entities.StoredUser{}, errors.New(RandomString())
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
	findUserFromToken func(string) (entities.StoredUser, error)
}

func (s *StubTokenStore) FindUserFromToken(token string) (entities.StoredUser, error) {
	if s.findUserFromToken != nil {
		return s.findUserFromToken(token)
	} else {
		return entities.StoredUser{}, token_store_contract.TokenNotFoundErr
	}
}
