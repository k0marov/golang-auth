package server

import (
	"auth/core/client_errors"
	"auth/domain/entities"
	"auth/values"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthService interface {
	Register(values.AuthData) (entities.Token, error)
	Login(values.AuthData) (entities.Token, error)
}

type AuthServer struct {
	service AuthService
}

func NewAuthServer(service AuthService) *AuthServer {
	return &AuthServer{service}
}

func (srv *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("contentType", "application/json")
	trimmedURL := strings.TrimRight(r.URL.String(), "/")
	if strings.HasSuffix(trimmedURL, "login") {
		srv.loginHandler(w, r)
	} else if strings.HasSuffix(trimmedURL, "register") {
		srv.registerHandler(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (srv *AuthServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	srv.baseHandler(w, r, srv.service.Login)
}

func (srv *AuthServer) registerHandler(w http.ResponseWriter, r *http.Request) {
	srv.baseHandler(w, r, srv.service.Register)
}

func (srv *AuthServer) baseHandler(w http.ResponseWriter, r *http.Request, callProperService func(values.AuthData) (entities.Token, error)) {
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
