package auth_integration_test

import (
	"auth/core/client_errors"
	"auth/core/crypto/bcrypt_hasher"
	"auth/data/store"
	"auth/data/store/db_file_interactor_impl"
	"auth/delivery/server"
	"auth/delivery/token_auth_middleware"
	"auth/domain/auth_service"
	"auth/domain/entities"
	. "auth/test_helpers"
	"auth/values"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthIntegration(t *testing.T) {
	tempDB, closeDB := CreateTempFile(t, "")
	defer closeDB()
	dbFileInteractor := db_file_interactor_impl.NewDBFileInteractor(tempDB)
	store, err := store.NewPersistentInMemoryFileStore(dbFileInteractor)
	AssertNoError(t, err)

	hasher := bcrypt_hasher.NewBcryptHasher(4)
	service := auth_service.NewAuthServiceImpl(store, hasher)
	server := server.NewAuthServer(service)

	baseAuthRequest := func(userData values.AuthData, endpoint string) *httptest.ResponseRecorder {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(userData)
		request := httptest.NewRequest(http.MethodPost, endpoint, body)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	requestLogin := func(userData values.AuthData) *httptest.ResponseRecorder {
		return baseAuthRequest(userData, "/login")
	}
	requestRegister := func(userData values.AuthData) *httptest.ResponseRecorder {
		return baseAuthRequest(userData, "/register")
	}
	middleware := token_auth_middleware.NewTokenAuthMiddleware(store).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(r.Context().Value("User"))
	}))
	requestMiddleware := func(token string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Add("Authorization", "Token "+token)
		response := httptest.NewRecorder()
		middleware.ServeHTTP(response, request)
		return response
	}

	username := "sam_komarov"
	password := "very_strong_password"
	passwordHashed, _ := bcrypt.GenerateFromPassword([]byte(password), 4)

	// check middleware with invalid token
	response := requestMiddleware("abracadabra")
	assertClientError(t, response, client_errors.AuthTokenInvalidError, http.StatusUnauthorized)

	// try to login when not registered yet
	response = requestLogin(values.AuthData{Username: username, Password: string(passwordHashed)})
	assertClientError(t, response, client_errors.InvalidCredentialsError, http.StatusBadRequest)

	// try to register with non-client hashed password
	response = requestRegister(values.AuthData{Username: username, Password: password})
	assertClientError(t, response, client_errors.UnhashedPasswordError, http.StatusBadRequest)

	// register with hashed password
	response = requestRegister(values.AuthData{Username: username, Password: string(passwordHashed)})
	registerToken := assertSuccessAndGetToken(t, response)

	// login into newly created account
	response = requestLogin(values.AuthData{Username: username, Password: string(passwordHashed)})
	loginToken := assertSuccessAndGetToken(t, response)
	Assert(t, loginToken, registerToken, "the token returned from login")

	// try to register another user with the same username
	response = requestRegister(values.AuthData{Username: username, Password: password})
	assertClientError(t, response, client_errors.UsernameAlreadyTakenError, http.StatusBadRequest)

	// check middleware with valid token
	response = requestMiddleware(loginToken.Token)
	assertSuccessAndValidUser(t, response, username)

}

func assertSuccessAndValidUser(t testing.TB, response *httptest.ResponseRecorder, username string) {
	t.Helper()
	Assert(t, response.Code, http.StatusOK, "response status code")
	var user entities.User
	json.NewDecoder(response.Body).Decode(&user)
	Assert(t, user.Username, user.Username, "user's username")
}

func assertSuccessAndGetToken(t testing.TB, response *httptest.ResponseRecorder) entities.Token {
	t.Helper()
	Assert(t, response.Code, http.StatusOK, "response status code")
	var token entities.Token
	json.NewDecoder(response.Result().Body).Decode(&token)
	return token
}

func assertClientError(t testing.TB, response *httptest.ResponseRecorder, error client_errors.ClientError, code int) {
	t.Helper()
	Assert(t, response.Code, code, "response status code")
	var gotError client_errors.ClientError
	json.NewDecoder(response.Result().Body).Decode(&gotError)
	Assert(t, gotError, error, "response client error")
}
