package models

import "github.com/k0marov/golang-auth/internal/domain/entities"

type UserModel struct {
	Id         int
	Username   string
	StoredPass string
	AuthToken  entities.Token
}
