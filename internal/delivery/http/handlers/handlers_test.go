package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/delivery/http/handlers"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	. "github.com/k0marov/golang-auth/internal/test_helpers"
	"github.com/k0marov/golang-auth/internal/values"
)

var goodAuthData = values.AuthData{
	Username: RandomString(),
	Password: RandomString(),
}

func TestRegisterHandler(t *testing.T) {
	baseTestHandler(t, handlers.NewRegisterHandler)
}

func TestLoginHandler(t *testing.T) {
	baseTestHandler(t, handlers.NewLoginHandler)
}

func baseTestHandler(t *testing.T, makeHandler func(handlers.AuthServiceMethod) http.HandlerFunc) {
	t.Helper()

	goodPostData := encodeAuthData(goodAuthData)
	makeRequest := func(postData string) *http.Request {
		return httptest.NewRequest(http.MethodOptions, "/url-should-not-be-used", bytes.NewReader([]byte(postData)))
	}

	actAndAssertJson := func(t testing.TB, handler http.Handler, postData string) *httptest.ResponseRecorder {
		request := makeRequest(postData)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		AssertJSON(t, response)
		return response
	}

	t.Run("should call service and return token if provided post data is ok and service call is successful", func(t *testing.T) {
		randomToken := entities.Token{Token: RandomString()}
		serviceMethod := func(authData values.AuthData) (entities.Token, error) {
			if authData == goodAuthData {
				return randomToken, nil
			}
			panic("called with unexpected arguments")
		}
		sut := makeHandler(serviceMethod)

		response := actAndAssertJson(t, sut, goodPostData)

		Assert(t, response.Code, 200, "status code")
		Assert(t, response.Body.String(), jsonString(randomToken), "response body")
	})
	t.Run("should return error if provided post data is not valid json", func(t *testing.T) {
		sut := makeHandler(nil) // service is nil, since it shouldn't be called
		response := actAndAssertJson(t, sut, "abracadabra")
		AssertHTTPError(t, response, client_errors.InvalidJsonError, http.StatusBadRequest)
	})
	t.Run("if business logic returns client error should return the same error", func(t *testing.T) {
		randomClientError := client_errors.ClientError{
			DetailCode:     RandomString(),
			ReadableDetail: RandomString(),
		}
		serviceMethod := func(values.AuthData) (entities.Token, error) {
			return entities.Token{}, randomClientError
		}
		sut := makeHandler(serviceMethod)

		response := actAndAssertJson(t, sut, goodPostData)
		AssertHTTPError(t, response, randomClientError, http.StatusBadRequest)
	})
	t.Run("if business logic returns not a client error should just return status code 500 and empty body", func(t *testing.T) {
		serviceMethod := func(values.AuthData) (entities.Token, error) {
			return entities.Token{}, errors.New(RandomString())
		}
		sut := makeHandler(serviceMethod)

		response := actAndAssertJson(t, sut, goodPostData)
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
