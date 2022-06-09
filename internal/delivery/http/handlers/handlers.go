package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/values"
)

type AuthServiceMethod = func(values.AuthData) (entities.Token, error)

func NewLoginHandler(login AuthServiceMethod) http.HandlerFunc {
	return newBaseHandler(login)
}

func NewRegisterHandler(register AuthServiceMethod) http.HandlerFunc {
	return newBaseHandler(register)
}

func newBaseHandler(callProperService AuthServiceMethod) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("contentType", "application/json")
		var postData values.AuthData
		err := json.NewDecoder(r.Body).Decode(&postData)
		if err != nil {
			throwHTTPError(w, client_errors.InvalidJsonError)
			return
		}
		token, err := callProperService(postData)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		json.NewEncoder(w).Encode(token)
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	if err, ok := err.(client_errors.ClientError); ok { // upcast to client error
		throwHTTPError(w, err)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func throwHTTPError(w http.ResponseWriter, httpErr client_errors.ClientError) {
	errorBuf := bytes.NewBuffer(nil)
	json.NewEncoder(errorBuf).Encode(httpErr)
	http.Error(w, errorBuf.String(), http.StatusBadRequest)
}
