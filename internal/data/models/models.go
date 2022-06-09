package models

import "internal/domain/entities"

type UserModel struct {
	Id         int
	Username   string
	StoredPass string
	AuthToken  entities.Token
}
