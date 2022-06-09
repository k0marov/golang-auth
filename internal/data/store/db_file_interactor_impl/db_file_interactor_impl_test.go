package db_file_interactor_impl_test

import (
	"io"
	"os"
	"sync"
	"testing"

	"internal/data/store/db_file_interactor_impl"
	. "internal/test_helpers"
)

func TestFileInteractor(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		testFileName, deleteFile := CreateTempFile(t, "")
		defer deleteFile()

		interactor := db_file_interactor_impl.NewDBFileInteractor(testFileName)
		users, err := interactor.ReadUsers()
		AssertNoError(t, err)
		Assert(t, len(users), 0, "length of parsed users")

		t.Run("populate file with 5 users, close it and then read", func(t *testing.T) {
			generatedUsers := GenerateRandomUserModels(5)
			for _, user := range generatedUsers {
				err := interactor.WriteUser(user)
				AssertNoError(t, err)
			}
			// emulate restarting the program
			interactor = db_file_interactor_impl.NewDBFileInteractor(testFileName)
			storedUsers, err := interactor.ReadUsers()
			AssertNoError(t, err)
			if !Assert(t, storedUsers, generatedUsers, "stored users") {
				testFile, _ := os.Open(testFileName)
				testFile.Seek(0, 0)
				fileContents, _ := io.ReadAll(testFile)
				t.Errorf("db file contents: %s", string(fileContents))
			}
		})
	})
	t.Run("should be safe for concurrent access", func(t *testing.T) {
		testFile, deleteFile := CreateTempFile(t, "")
		defer deleteFile()
		interactor := db_file_interactor_impl.NewDBFileInteractor(testFile)

		wantedCount := 1000

		var wg sync.WaitGroup
		wg.Add(wantedCount)

		for i := 0; i < wantedCount; i++ {
			go func() {
				err := interactor.WriteUser(GenerateRandomUserModel())
				AssertNoError(t, err)
				wg.Done()
			}()
		}
		wg.Wait()

		users, err := interactor.ReadUsers()
		AssertNoError(t, err)
		Assert(t, len(users), wantedCount, "number of created users")
	})
	t.Run("error cases", func(t *testing.T) {
		// TODO: Try to test file reading/writing (maybe using afero)
	})
}
