package store_test

import (
	"errors"
	"sync"
	"testing"

	"internal/data/models"
	"internal/data/store"
	"internal/domain/auth_store_contract"
	"internal/domain/entities"
	"internal/domain/token_store_contract"
	. "internal/test_helpers"
)

func TestPersistentInMemoryFileStore(t *testing.T) {
	t.Run("CreateUser() id generation", func(t *testing.T) {
		createRandomUser := func(t testing.TB, sutStore *store.PersistentInMemoryFileStore) int {
			createdUser, err := sutStore.CreateUser(RandomString(), RandomString(), entities.Token{Token: RandomString()})
			AssertNoError(t, err)
			return createdUser.Id
		}
		t.Run("should generate auto incrementing ids starting from 1", func(t *testing.T) {
			fileInteractor := &StubDBFileInteractor{}
			sutStore, err := store.NewPersistentInMemoryFileStore(fileInteractor)
			AssertNoError(t, err)

			const testCount = 10000

			for i := 1; i <= testCount; i++ {
				AssertFatal(t, createRandomUser(t, sutStore), i, "generated id")
			}

			t.Run("should continue incrementing each new id properly after a restart", func(t *testing.T) {
				// simulate restart of the application (fileInteractor is the same since it is initialized with a persistent file)
				sutStore, err = store.NewPersistentInMemoryFileStore(fileInteractor)
				AssertNoError(t, err)

				for i := testCount + 1; i <= testCount*2; i++ {
					AssertFatal(t, createRandomUser(t, sutStore), i, "generated id")
				}
			})
		})
		t.Run("id generation should be safe for concurrent usage", func(t *testing.T) {
			fileInteractor := &StubDBFileInteractor{}
			sutStore, err := store.NewPersistentInMemoryFileStore(fileInteractor)
			AssertNoError(t, err)

			wantedCount := 1000

			var wg sync.WaitGroup
			wg.Add(wantedCount)

			generatedIds := make([]int, wantedCount)
			for i := 0; i < wantedCount; i++ {
				go func(i int) {
					newId := createRandomUser(t, sutStore)
					generatedIds[i] = newId
					wg.Done()
				}(i)
			}
			wg.Wait()

			AssertUniqueCount(t, generatedIds, wantedCount)
		})
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
		t.Run("CreateUser() should return error if write failed (and do not save any user)", func(t *testing.T) {
			errorFileInteractor := &ErrorDBFileInteractor{ThrowOnRead: false, ThrowOnWrite: true}
			store, err := store.NewPersistentInMemoryFileStore(errorFileInteractor)
			AssertNoError(t, err)

			randomUser := GenerateRandomUser()

			_, err = store.CreateUser(randomUser.Username, randomUser.Password, randomUser.Token)
			AssertSomeError(t, err)

			assertUserNotInStore(t, store, randomUser)
		})
	})
}

func createUsers(t testing.TB, store *store.PersistentInMemoryFileStore, newUsers []RandomUser) (newIds []int) {
	t.Helper()
	for _, newUser := range newUsers {
		createdUser, err := store.CreateUser(newUser.Username, newUser.Password, newUser.Token)
		AssertNoError(t, err)
		newIds = append(newIds, createdUser.Id)
	}
	return
}

func assertUsersInStore(t testing.TB, store *store.PersistentInMemoryFileStore, users []RandomUser, ids []int) {
	t.Helper()
	for i, newUser := range users {
		assertUserInStore(t, store, newUser, ids[i])
	}
}

func assertUserInStore(t testing.TB, store *store.PersistentInMemoryFileStore, user RandomUser, id int) {
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
func assertUserEqual(t testing.TB, userInStore models.UserModel, want RandomUser) {
	t.Helper()
	Assert(t, userInStore.Username, want.Username, "username")
	Assert(t, userInStore.StoredPass, want.Password, "password")
	Assert(t, userInStore.AuthToken, want.Token, "token")
}

type StubDBFileInteractor struct {
	users []models.UserModel
	mu    sync.Mutex
}

func (s *StubDBFileInteractor) ReadUsers() ([]models.UserModel, error) {
	return s.users, nil
}
func (s *StubDBFileInteractor) WriteUser(user models.UserModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = append(s.users, user)
	return nil
}

type ErrorDBFileInteractor struct {
	ThrowOnRead  bool
	ThrowOnWrite bool
}

func (e *ErrorDBFileInteractor) ReadUsers() ([]models.UserModel, error) {
	var err error
	if e.ThrowOnRead {
		err = errors.New(RandomString())
	} else {
		err = nil
	}
	return []models.UserModel{}, err
}
func (e *ErrorDBFileInteractor) WriteUser(user models.UserModel) error {
	if e.ThrowOnWrite {
		return errors.New(RandomString())
	}
	return nil
}
