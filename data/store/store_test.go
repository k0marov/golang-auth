package store_test

import (
	"auth/data/store"
	"auth/domain/auth_store_contract"
	"auth/domain/entities"
	"auth/domain/token_store_contract"
	. "auth/test_helpers"
	"errors"
	"testing"
)

type StubDBFileInteractor struct {
	users []entities.StoredUser
}

func (s *StubDBFileInteractor) ReadUsers() ([]entities.StoredUser, error) {
	return s.users, nil
}
func (s *StubDBFileInteractor) WriteUser(user entities.StoredUser) error {
	s.users = append(s.users, user)
	return nil
}

type FakeDBFileInteractor struct{}

func (f *FakeDBFileInteractor) ReadUsers() ([]entities.StoredUser, error) {
	return []entities.StoredUser{}, nil
}
func (f *FakeDBFileInteractor) WriteUser(user entities.StoredUser) error {
	return nil
}

type ErrorDBFileInteractor struct {
	ThrowOnRead  bool
	ThrowOnWrite bool
}

func (e *ErrorDBFileInteractor) ReadUsers() ([]entities.StoredUser, error) {
	var err error
	if e.ThrowOnRead {
		err = errors.New(RandomString())
	} else {
		err = nil
	}
	return []entities.StoredUser{}, err
}
func (e *ErrorDBFileInteractor) WriteUser(user entities.StoredUser) error {
	if e.ThrowOnWrite {
		return errors.New(RandomString())
	}
	return nil
}

func TestPersistentInMemoryFileStore(t *testing.T) {
	t.Run("CreateUser() should generate random and unique IDs", func(t *testing.T) {
		store, err := store.NewPersistentInMemoryFileStore(&FakeDBFileInteractor{})
		AssertNoError(t, err)

		const testCount = 10000

		generatedIds := []string{}
		for i := 0; i < testCount; i++ {
			createdUser, err := store.CreateUser(RandomString(), RandomString(), entities.Token{Token: RandomString()})
			AssertNoError(t, err)
			generatedIds = append(generatedIds, createdUser.Id)
		}
		AssertUniqueCount(t, generatedIds, testCount)
	})
	t.Run("in memory works", func(t *testing.T) {
		fileInteractor := &StubDBFileInteractor{}
		sutStore, err := store.NewPersistentInMemoryFileStore(fileInteractor)
		AssertNoError(t, err)

		newUsers := GenerateRandomUsers(5)
		newIds := createUsers(t, sutStore, newUsers) // this uses CreateUser()
		// this uses UserExists(), FindUser(), FindUserFromToken()
		assertUsersInStore(t, sutStore, newUsers, newIds)

		// a user, which we have not added yet, should not "exist" (by the definition of READ methods) in a database
		randomNotStoredUser := GenerateRandomUser()
		assertUserNotInStore(t, sutStore, randomNotStoredUser)

		t.Run("persistence works", func(t *testing.T) {
			// simulate restart of the application (fileInteractor is the same since it is initialized with a persistent file)
			sutStore, err = store.NewPersistentInMemoryFileStore(fileInteractor)
			AssertNoError(t, err)

			assertUsersInStore(t, sutStore, newUsers, newIds)

			anotherNewUsers := GenerateRandomUsers(3)
			anotherNewIds := createUsers(t, sutStore, anotherNewUsers)
			assertUsersInStore(t, sutStore, newUsers, newIds)
			assertUsersInStore(t, sutStore, anotherNewUsers, anotherNewIds)
		})
	})
	t.Run("test error handling", func(t *testing.T) {
		t.Run("constructor should return error if read failed", func(t *testing.T) {
			errorFileInteractor := &ErrorDBFileInteractor{ThrowOnRead: true, ThrowOnWrite: false}
			_, err := store.NewPersistentInMemoryFileStore(errorFileInteractor)
			AssertSomeError(t, err)
		})
		t.Run("CreateUser() should return error if write failed", func(t *testing.T) {
			errorFileInteractor := &ErrorDBFileInteractor{ThrowOnRead: false, ThrowOnWrite: true}
			store, err := store.NewPersistentInMemoryFileStore(errorFileInteractor)
			AssertNoError(t, err)
			_, err = store.CreateUser(RandomString(), RandomString(), entities.Token{Token: RandomString()})
			AssertSomeError(t, err)
		})
	})
}

func createUsers(t testing.TB, store *store.PersistentInMemoryFileStore, newUsers []RandomUser) (newIds []string) {
	t.Helper()
	for _, newUser := range newUsers {
		createdUser, err := store.CreateUser(newUser.Username, newUser.Password, newUser.Token)
		AssertNoError(t, err)
		newIds = append(newIds, createdUser.Id)
	}
	return
}

func assertUsersInStore(t testing.TB, store *store.PersistentInMemoryFileStore, users []RandomUser, ids []string) {
	t.Helper()
	for i, newUser := range users {
		assertUserInStore(t, store, newUser, ids[i])
	}
}

func assertUserInStore(t testing.TB, store *store.PersistentInMemoryFileStore, user RandomUser, id string) {
	t.Helper()
	// UserExists()
	Assert(t, store.UserExists(user.Username), true, "UserExists()")
	// FindUser()
	userInStore, err := store.FindUser(user.Username)
	AssertNoError(t, err)
	assertUserEqual(t, userInStore, user)
	Assert(t, userInStore.Id, id, "id of created user")
	// FindUserFromToken()
	userInStore, err = store.FindUserFromToken(user.Token.Token)
	AssertNoError(t, err)
	assertUserEqual(t, userInStore, user)
	Assert(t, userInStore.Id, id, "id of created user")
}
func assertUserNotInStore(t testing.TB, store *store.PersistentInMemoryFileStore, user RandomUser) {
	t.Helper()
	Assert(t, store.UserExists(user.Username), false, "UserExists()")
	_, err := store.FindUser(user.Username)
	AssertError(t, err, auth_store_contract.UserNotFoundErr)
	_, err = store.FindUserFromToken(user.Token.Token)
	AssertError(t, err, token_store_contract.TokenNotFoundErr)
}
func assertUserEqual(t testing.TB, userInStore entities.StoredUser, want RandomUser) {
	t.Helper()
	Assert(t, userInStore.Username, want.Username, "username")
	Assert(t, userInStore.StoredPass, want.Password, "password")
	Assert(t, userInStore.AuthToken, want.Token, "token")
}
