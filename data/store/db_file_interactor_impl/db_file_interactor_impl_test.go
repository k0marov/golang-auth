package db_file_interactor_impl_test

import (
	"auth/data/store/db_file_interactor_impl"
	"auth/domain/entities"
	. "auth/test_helpers"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
)

func TestFileInteractor(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		testFile, deleteFile := CreateTempFile(t, "")

		interactor := db_file_interactor_impl.NewDBFileInteractor(testFile)
		users, err := interactor.ReadUsers()
		AssertNoError(t, err)
		Assert(t, len(users), 0, "length of parsed users")

		t.Run("populate file with 5 users, close it and then read", func(t *testing.T) {
			generatedUsers := generateRandomStoredUsers(5)
			for _, user := range generatedUsers {
				err := interactor.WriteUser(user)
				AssertNoError(t, err)
			}
			// emulate restarting the program
			interactor = db_file_interactor_impl.NewDBFileInteractor(testFile)
			storedUsers, err := interactor.ReadUsers()
			AssertNoError(t, err)
			if !Assert(t, storedUsers, generatedUsers, "stored users") {
				testFile.Seek(0, 0)
				fileContents, _ := io.ReadAll(testFile)
				t.Errorf("db file contents: %s", string(fileContents))
			}
		})
		deleteFile()
	})
	t.Run("error cases", func(t *testing.T) {
		t.Run("ReadUsers()", func(t *testing.T) {
			assertError := func(errorDBFile *ErrorDBFile) {
				interactor := db_file_interactor_impl.NewDBFileInteractor(errorDBFile)
				_, err := interactor.ReadUsers()
				AssertSomeError(t, err)
			}
			t.Run("should return error if Seek() returns an error", func(t *testing.T) {
				assertError(&ErrorDBFile{throwOnSeek: true})
			})
			t.Run("should return error if Read() returns an error", func(t *testing.T) {
				assertError(&ErrorDBFile{throwOnRead: true})
			})
		})
		t.Run("WriteUser()", func(t *testing.T) {
			// not tested
		})
	})
	t.Run("should be safe for concurrent access", func(t *testing.T) {
		testFile, closeFile := CreateTempFile(t, "")
		defer closeFile()
		interactor := db_file_interactor_impl.NewDBFileInteractor(testFile)

		wantedCount := 300

		var wg sync.WaitGroup
		wg.Add(wantedCount)

		for i := 0; i < wantedCount; i++ {
			go func() {
				err := interactor.WriteUser(generateRandomStoredUser())
				AssertNoError(t, err)
				wg.Done()
			}()
		}
		wg.Wait()

		users, err := interactor.ReadUsers()
		AssertNoError(t, err)
		Assert(t, len(users), wantedCount, "number of created users")
	})
}

func generateRandomStoredUser() entities.StoredUser {
	return generateRandomStoredUsers(1)[0]
}

func generateRandomStoredUsers(count int) (storedUsers []entities.StoredUser) {
	randomUsers := GenerateRandomUsers(count)
	for _, user := range randomUsers {
		storedUsers = append(storedUsers, entities.StoredUser{
			Id:         RandomString(),
			Username:   user.Username,
			StoredPass: user.Password,
			AuthToken:  user.Token,
		})
	}
	return
}

type ErrorDBFile struct {
	throwOnSeek bool
	throwOnRead bool
}

func (e *ErrorDBFile) Read(p []byte) (n int, err error) {
	if e.throwOnRead {
		return 1, errors.New(RandomString())
	}
	return 1, nil
}

func (e *ErrorDBFile) Seek(offset int64, whence int) (int64, error) {
	if e.throwOnSeek {
		return 1, errors.New(RandomString())
	}
	return 1, nil
}

func (e *ErrorDBFile) Write(p []byte) (n int, err error) {
	return 1, nil
}
func (e *ErrorDBFile) Truncate(size int64) error {
	return nil
}

func (e *ErrorDBFile) Stat() (os.FileInfo, error) {
	return nil, nil
}
