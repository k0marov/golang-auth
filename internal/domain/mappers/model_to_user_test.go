package mappers_test

import (
	"testing"

	"github.com/k0marov/golang-auth/internal/data/models"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	"github.com/k0marov/golang-auth/internal/domain/mappers"

	. "github.com/k0marov/golang-auth/internal/test_helpers"
)

func TestModelToUser(t *testing.T) {
	cases := []struct {
		model  models.UserModel
		entity entities.User
	}{
		{models.UserModel{
			Id:         42,
			Username:   "John",
			StoredPass: "abc",
			AuthToken:  entities.Token{Token: "xyz"},
		}, entities.User{
			Id:       "42",
			Username: "John",
		}},
		{models.UserModel{
			Id:         33,
			Username:   "Jack",
			StoredPass: "abc",
			AuthToken:  entities.Token{Token: "xyz"},
		}, entities.User{
			Id:       "33",
			Username: "Jack",
		}},
	}

	for _, c := range cases {
		Assert(t, mappers.ModelToUser(c.model), c.entity, "converted entity")
	}
}
