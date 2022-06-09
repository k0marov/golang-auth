package mappers

import (
	"strconv"

	"internal/data/models"
	"internal/domain/entities"
)

func ModelToUser(model models.UserModel) entities.User {
	return entities.User{
		Id:       strconv.Itoa(model.Id),
		Username: model.Username,
	}
}
