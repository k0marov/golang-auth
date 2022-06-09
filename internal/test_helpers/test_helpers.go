package test_helpers

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"internal/core/client_errors"
	"internal/data/models"
	"internal/domain/entities"
)

func Assert[T any](t testing.TB, got, want T, description string) bool {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s is not right: got '%v', want '%v'", description, got, want)
		return false
	}
	return true
}

func AssertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Errorf("expected no error but got %v", got)
	}
}
func AssertError(t testing.TB, got error, want error) {
	t.Helper()
	if got != want {
		t.Errorf("expected error %v, but got %v", want, got)
	}
}
func AssertSomeError(t testing.TB, got error) {
	t.Helper()
	if got == nil {
		t.Error("expected an error, but got nil")
	}
}

func AssertFatal[T comparable](t testing.TB, got, want T, description string) {
	t.Helper()
	if !Assert(t, got, want, description) {
		t.Fatal()
	}
}

func AssertNotNil[T comparable](t testing.TB, got T, description string) {
	t.Helper()
	var nilT T
	if got == nilT {
		t.Errorf("expected %s to be non nil, but got nil", description)
	}
}

func AssertHTTPError(t testing.TB, response *httptest.ResponseRecorder, err client_errors.ClientError, statusCode int) {
	t.Helper()
	var got client_errors.ClientError
	json.NewDecoder(response.Body).Decode(&got)

	Assert(t, got, err, "error response")
	Assert(t, response.Code, statusCode, "status code")
}

func AssertJSON(t testing.TB, response *httptest.ResponseRecorder) {
	t.Helper()
	Assert(t, response.Result().Header.Get("contentType"), "application/json", "response content type")
}

func CreateTempFile(t testing.TB, initialData string) (fileName string, deleteFile func()) {
	t.Helper()

	tmpfile, err := ioutil.TempFile(".", "db")
	name := tmpfile.Name()
	tmpfile.Write([]byte(initialData))
	if err != nil {
		t.Fatalf("could not create temp file %v", err)
	}
	tmpfile.Close()

	removeFile := func() {
		os.Remove(tmpfile.Name())
	}

	return name, removeFile
}

type RandomUser struct {
	Username string
	Password string
	Token    entities.Token
}

func GenerateRandomUsers(count int) []RandomUser {
	result := []RandomUser{}
	for i := 0; i < count; i++ {
		result = append(result, GenerateRandomUser())
	}
	return result
}

func GenerateRandomUser() RandomUser {
	return RandomUser{
		Username: RandomString(),
		Password: RandomString(),
		Token:    entities.Token{Token: RandomString()},
	}
}

func GenerateRandomUserModel() models.UserModel {
	return GenerateRandomUserModels(1)[0]
}

func GenerateRandomUserModels(count int) (storedUsers []models.UserModel) {
	randomUsers := GenerateRandomUsers(count)
	for _, user := range randomUsers {
		storedUsers = append(storedUsers, models.UserModel{
			Id:         RandomInt(),
			Username:   user.Username,
			StoredPass: user.Password,
			AuthToken:  user.Token,
		})
	}
	return
}

func AssertUniqueCount[T comparable](t testing.TB, slice []T, want int) {
	t.Helper()
	unique := []T{}
	for _, val := range slice {
		if !CheckInSlice(val, unique) {
			unique = append(unique, val)
		}
	}
	Assert(t, len(unique), want, "number of unique elements")
}

func CheckInSlice[T comparable](elem T, slice []T) bool {
	for _, sliceElem := range slice {
		if sliceElem == elem {
			return true
		}
	}
	return false
}

func RandomInt() int {
	return rand.Intn(100)
}

func RandomString() string {
	str := ""
	for i := 0; i < 2; i++ {
		str += words[rand.Intn(len(words))] + "_"
	}
	return str
}

var words = []string{"the", "be", "to", "of", "and", "a", "in", "that", "have", "I", "it", "for", "not", "on", "with", "he", "as", "you", "do", "at", "this", "but", "his", "by", "from", "they", "we", "say", "her", "she", "or", "an", "will", "my", "one", "all", "would", "there", "their", "what", "so", "up", "out", "if", "about", "who", "get", "which", "go", "me", "when", "make", "can", "like", "time", "no", "just", "him", "know", "take", "people", "into", "year", "your", "good", "some", "could", "them", "see", "other", "than", "then", "now", "look", "only", "come", "its", "over", "think", "also", "back", "after", "use", "two", "how", "our", "work", "first", "well", "way", "even", "new", "want", "because", "any", "these", "give", "day", "most", "us"}
