package mappers

import (
	"strconv"

	"github.com/k0marov/golang-auth/internal/data/models"
	"github.com/k0marov/golang-auth/internal/domain/entities"
)

func ModelToUser(model models.UserModel) entities.User {
	return entities.User{
		Id:       strconv.Itoa(model.Id),
		Username: model.Username,
	}
}
