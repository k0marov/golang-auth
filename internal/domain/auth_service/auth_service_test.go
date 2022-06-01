package auth_service_test

import (
	"errors"
	"testing"

	"github.com/k0marov/golang-auth/internal/core/client_errors"
	"github.com/k0marov/golang-auth/internal/domain/auth_service"
	"github.com/k0marov/golang-auth/internal/domain/auth_store_contract"
	"github.com/k0marov/golang-auth/internal/domain/entities"
	. "github.com/k0marov/golang-auth/internal/test_helpers"
	"github.com/k0marov/golang-auth/internal/values"
)

type StubAuthStore struct {
	userExists func(string) bool
	createUser func(string, string, entities.Token) (entities.StoredUser, error)
	findUser   func(string) (entities.StoredUser, error)
}

func (s *StubAuthStore) UserExists(username string) bool {
	if s.userExists != nil {
		return s.userExists(username)
	}
	return false
}

func (s *StubAuthStore) CreateUser(username string, hashedPassword string, token entities.Token) (entities.StoredUser, error) {
	if s.createUser != nil {
		return s.createUser(username, hashedPassword, token)
	}
	return entities.StoredUser{}, nil
}

func (s *StubAuthStore) FindUser(username string) (entities.StoredUser, error) {
	if s.findUser != nil {
		return s.findUser(username)
	}
	return entities.StoredUser{}, auth_store_contract.UserNotFoundErr
}

type StubHasher struct {
	isHashed func(string) bool
	hash     func(string) (string, error)
	compare  func(string, string) bool
}

func (s StubHasher) IsHashed(pass string) bool {
	if s.isHashed == nil {
		return true
	}
	return s.isHashed(pass)
}
func (s StubHasher) Hash(pass string) (string, error) {
	if s.hash == nil {
		return pass, nil
	}
	return s.hash(pass)
}
func (s StubHasher) Compare(pass, hashedPass string) bool {
	if s.compare == nil {
		return true
	}
	return s.compare(pass, hashedPass)
}

var dummyStore = &StubAuthStore{}
var dummyHasher = &StubHasher{}

func TestAuthService_Register(t *testing.T) {
	t.Run("should check if user with provided username already exists in the store", func(t *testing.T) {
		// arrange
		takenUsername := RandomString()
		newUsername := RandomString()
		store := &StubAuthStore{
			userExists: func(username string) bool {
				if username == takenUsername {
					return true
				}
				return false
			},
		}
		service := auth_service.NewAuthServiceImpl(store, dummyHasher)

		t.Run("happy case", func(t *testing.T) {
			_, err := service.Register(values.AuthData{
				Username: newUsername,
				Password: RandomString(),
			})
			AssertNoError(t, err)
		})

		t.Run("error case (username already taken)", func(t *testing.T) {
			// error case
			_, err := service.Register(values.AuthData{
				Username: takenUsername,
				Password: RandomString(),
			})
			AssertError(t, err, client_errors.UsernameAlreadyTakenError)
		})

	})
	t.Run("should check if password is client-side hashed", func(t *testing.T) {
		// arrange
		pass := RandomString()
		hashedPass := RandomString()
		hasher := StubHasher{
			isHashed: func(pass string) bool {
				return pass == hashedPass
			},
		}
		service := auth_service.NewAuthServiceImpl(dummyStore, hasher)

		// happy case
		_, err := service.Register(values.AuthData{
			Username: RandomString(),
			Password: hashedPass,
		})
		AssertNoError(t, err)
		// error case
		_, err = service.Register(values.AuthData{
			Username: RandomString(),
			Password: pass,
		})
		Assert[error](t, err, client_errors.UnhashedPasswordError, "expected error")
	})
	t.Run("should check if username is valid (contains proper characters)", func(t *testing.T) {
		AssertUniqueCount(t, []byte(auth_service.ValidUsernameChars), 10+26*2+1)
		service := auth_service.NewAuthServiceImpl(dummyStore, dummyHasher)
		cases := []struct {
			username string
			valid    bool
		}{
			{"", false},
			{"a", true},
			{"asdf", true},
			{"asdF", true},
			{"aSdf_asdkfljas", true},
			{"sadklfjklasjdfkjsdlfjskldjfkljasdklfjkasjdf", false}, // too long
			{"$adS&&..'", false},
			{"_asdf", false},
			{"asdf8348", true},
			{"123sasdf", true},
		}

		for _, c := range cases {
			t.Run(c.username, func(t *testing.T) {
				_, err := service.Register(values.AuthData{
					Username: c.username,
					Password: RandomString(),
				})
				if c.valid {
					AssertNoError(t, err)
				} else {
					AssertError(t, err, client_errors.UsernameInvalidError)
				}
			})
		}
	})
	t.Run("should create a new user in the store (with password hashed second time) and return the right token if all checks have passed", func(t *testing.T) {
		type createArgs struct {
			username string
			password string
			token    entities.Token
		}
		t.Run("happy case", func(t *testing.T) {
			createCalledWith := []createArgs{}
			store := &StubAuthStore{
				createUser: func(username string, password string, token entities.Token) (entities.StoredUser, error) {
					createCalledWith = append(createCalledWith, createArgs{username, password, token})
					return entities.StoredUser{}, nil
				},
			}

			rightHashedPass := RandomString()
			rightUsername := RandomString()
			hasher := StubHasher{
				hash: func(string) (string, error) { return rightHashedPass, nil },
			}
			service := auth_service.NewAuthServiceImpl(store, hasher)

			token, err := service.Register(values.AuthData{
				Username: rightUsername,
				Password: RandomString(),
			})
			AssertNoError(t, err)
			Assert(t, len(createCalledWith), 1, "number of times CreateUser was called")
			Assert(t, createCalledWith[0], createArgs{rightUsername, rightHashedPass, token}, "CreateUser args")
		})
		t.Run("hasher returns an error (do not create new user)", func(t *testing.T) {
			createCalls := 0
			store := &StubAuthStore{
				createUser: func(username string, password string, token entities.Token) (entities.StoredUser, error) {
					createCalls++
					return entities.StoredUser{}, nil
				},
			}

			hasherErr := errors.New(RandomString())
			hasher := StubHasher{
				hash: func(string) (string, error) { return "", hasherErr },
			}
			service := auth_service.NewAuthServiceImpl(store, hasher)

			_, err := service.Register(values.AuthData{
				Username: RandomString(),
				Password: RandomString(),
			})

			AssertSomeError(t, err)
			Assert(t, createCalls, 0, "no users should be created")
		})
		t.Run("store returns an error", func(t *testing.T) {
			store := &StubAuthStore{
				createUser: func(username, password string, token entities.Token) (entities.StoredUser, error) {
					return entities.StoredUser{}, errors.New(RandomString())
				},
			}
			hasher := StubHasher{}
			service := auth_service.NewAuthServiceImpl(store, hasher)

			_, err := service.Register(values.AuthData{
				Username: RandomString(),
				Password: RandomString(),
			})
			AssertSomeError(t, err)
		})
	})
	t.Run("the generated token should be unique", func(t *testing.T) {
		wantedCount := 10000
		service := auth_service.NewAuthServiceImpl(dummyStore, dummyHasher)

		tokens := []entities.Token{}
		for i := 0; i < wantedCount; i++ {
			result, err := service.Register(values.AuthData{
				Username: RandomString(),
				Password: RandomString(),
			})
			AssertNoError(t, err)
			tokens = append(tokens, result)
		}

		AssertUniqueCount(t, tokens, wantedCount)
	})
}

func TestAuthService_Login(t *testing.T) {
	existingUsername := RandomString()
	hisPass := RandomString()
	hisPassHashed := RandomString()
	hisToken := entities.Token{Token: RandomString()}
	store := &StubAuthStore{
		findUser: func(username string) (entities.StoredUser, error) {
			if username == existingUsername {
				return entities.StoredUser{
					Username:   existingUsername,
					StoredPass: hisPassHashed,
					AuthToken:  hisToken,
				}, nil
			} else {
				return entities.StoredUser{}, auth_store_contract.UserNotFoundErr
			}
		},
	}
	t.Run("should call store to find user with provided username", func(t *testing.T) {
		service := auth_service.NewAuthServiceImpl(store, dummyHasher)

		t.Run("happy case (user found)", func(t *testing.T) {
			_, err := service.Login(values.AuthData{
				Username: existingUsername,
				Password: hisPass,
			})

			AssertNoError(t, err)
		})
		t.Run("error case (there is no user with such username)", func(t *testing.T) {
			_, err := service.Login(values.AuthData{Username: RandomString(), Password: RandomString()})
			AssertError(t, err, client_errors.InvalidCredentialsError)
		})
	})
	t.Run("should compare given password with password from the store", func(t *testing.T) {
		service := auth_service.NewAuthServiceImpl(store, &StubHasher{
			compare: func(pass, storedPass string) bool {
				if pass == hisPass && storedPass == hisPassHashed {
					return true
				}
				return false
			},
		})
		t.Run("happy case (passwords match)", func(t *testing.T) {
			_, err := service.Login(values.AuthData{
				Username: existingUsername,
				Password: hisPass,
			})
			AssertNoError(t, err)
		})
		t.Run("error case (passwords don't match)", func(t *testing.T) {
			_, err := service.Login(values.AuthData{
				Username: existingUsername,
				Password: "abracadabra",
			})
			AssertError(t, err, client_errors.InvalidCredentialsError)
		})
	})
	t.Run("should return the token from the store if credentials are valid", func(t *testing.T) {
		service := auth_service.NewAuthServiceImpl(store, dummyHasher)

		token, err := service.Login(values.AuthData{
			Username: existingUsername,
			Password: hisPass,
		})
		AssertNoError(t, err)
		Assert(t, token, hisToken, "the returned token")
	})
}
