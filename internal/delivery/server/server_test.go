package server_test

import (
	"auth/internal/core/client_errors"
	"auth/internal/delivery/server"
	"auth/internal/domain/entities"
	. "auth/internal/test_helpers"
	"auth/internal/values"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type spyAuthService struct {
	returnedToken entities.Token
	returnedError error
	registerCalls []values.AuthData
	loginCalls    []values.AuthData
}

func (s *spyAuthService) Register(user values.AuthData) (entities.Token, error) {
	s.registerCalls = append(s.registerCalls, user)
	return s.returnedToken, s.returnedError
}
func (s *spyAuthService) Login(user values.AuthData) (entities.Token, error) {
	s.loginCalls = append(s.loginCalls, user)
	return s.returnedToken, s.returnedError
}

var dummyAuthService = &spyAuthService{}

var goodAuthData = values.AuthData{
	Username: RandomString(),
	Password: RandomString(),
}

func TestServer_NotRegisterOrAuth(t *testing.T) {
	t.Run("should return 404 if request is neither to /register nor /login", func(t *testing.T) {
		authService := spyAuthService{}
		srv := server.NewAuthServer(&authService)

		response := httptest.NewRecorder()
		srv.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/posts/", nil))

		Assert(t, response.Result().StatusCode, http.StatusNotFound, "response status code")
		Assert(t, len(authService.loginCalls), 0, "number of login calls")
		Assert(t, len(authService.registerCalls), 0, "number of register calls")
	})
}

func TestServer_Register(t *testing.T) {
	makeRequest := func(postData string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(postData))
	}
	checkServiceCalls := func(spy *spyAuthService) bool {
		return len(spy.loginCalls) == 0 && len(spy.registerCalls) == 1 && spy.registerCalls[0] == goodAuthData
	}
	baseTestAuthMethod(t, makeRequest, checkServiceCalls)
}

func TestServer_Login(t *testing.T) {
	makeRequest := func(postData string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(postData))
	}
	checkServiceCalls := func(spy *spyAuthService) bool {
		return len(spy.registerCalls) == 0 && len(spy.loginCalls) == 1 && spy.loginCalls[0] == goodAuthData
	}
	baseTestAuthMethod(t, makeRequest, checkServiceCalls)
}

func baseTestAuthMethod(t *testing.T, makeRequest func(postData string) *http.Request, checkServiceCalls func(*spyAuthService) bool) {
	t.Helper()

	goodPostData := encodeAuthData(goodAuthData)

	actAndAssertJson := func(t testing.TB, srv *server.AuthServer, postData string) *httptest.ResponseRecorder {
		request := makeRequest(postData)
		response := httptest.NewRecorder()
		srv.ServeHTTP(response, request)

		AssertJSON(t, response)
		return response
	}

	t.Run("should call service and return token if provided post data is ok and service call is successful", func(t *testing.T) {
		randomToken := entities.Token{Token: RandomString()}
		authService := &spyAuthService{
			returnedToken: randomToken,
		}
		srv := server.NewAuthServer(authService)

		response := actAndAssertJson(t, srv, goodPostData)

		if !checkServiceCalls(authService) {
			t.Errorf("didn't trigger right service calls. login calls: %v, register calls: %v", authService.loginCalls, authService.registerCalls)
		}
		Assert(t, response.Code, 200, "status code")
		Assert(t, response.Body.String(), jsonString(randomToken), "response body")
	})
	t.Run("should return error if provided post data is not valid json", func(t *testing.T) {
		srv := server.NewAuthServer(dummyAuthService)

		response := actAndAssertJson(t, srv, "abracadabra")

		AssertHTTPError(t, response, client_errors.InvalidJsonError, http.StatusBadRequest)
	})
	t.Run("if business logic returns client error should return the same error", func(t *testing.T) {
		randomClientError := client_errors.ClientError{
			DetailCode:     RandomString(),
			ReadableDetail: RandomString(),
		}
		authService := spyAuthService{returnedError: randomClientError}
		srv := server.NewAuthServer(&authService)

		response := actAndAssertJson(t, srv, goodPostData)

		AssertHTTPError(t, response, randomClientError, http.StatusBadRequest)
	})
	t.Run("if business logic returns not a client error should just return status code 500 and empty body", func(t *testing.T) {
		randomServerError := errors.New(RandomString())
		authService := spyAuthService{returnedError: randomServerError}
		srv := server.NewAuthServer(&authService)

		response := actAndAssertJson(t, srv, goodPostData)

		Assert(t, response.Code, http.StatusInternalServerError, "status code")
		Assert(t, response.Body.String(), "", "response body")
	})
}

func encodeAuthData(data values.AuthData) string {
	post := bytes.NewBuffer(nil)
	json.NewEncoder(post).Encode(data)
	return post.String()
}

func jsonString[T any](obj T) string {
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(obj)
	return buf.String()
}
