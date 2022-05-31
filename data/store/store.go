package store

import (
	"auth/domain/auth_store_contract"
	"auth/domain/entities"
	"auth/domain/token_store_contract"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type DBFileInteractor interface {
	ReadUsers() ([]entities.StoredUser, error)
	WriteUser(entities.StoredUser) error
}

// An in-memory database is used here for 2 reasons:
// 1. This database is accessed on nearly every request (see TokenAuthMiddleware), so speed is needed
// 2. The schema is quite light on memory - 50 MB of RAM is enough to hold 100 000+ users
type PersistentInMemoryFileStore struct {
	fileInteractor DBFileInteractor
	usernameToUser map[string]*entities.StoredUser
	tokenToUser    map[string]*entities.StoredUser
	users          []entities.StoredUser
}

func NewPersistentInMemoryFileStore(fileInteractor DBFileInteractor) (*PersistentInMemoryFileStore, error) {
	users, err := fileInteractor.ReadUsers()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Got an error while reading users from file interactor: %v", err))
	}
	usernameToUser := make(map[string]*entities.StoredUser)
	tokenToUser := make(map[string]*entities.StoredUser)

	for i := range users {
		usernameToUser[users[i].Username] = &users[i]
		tokenToUser[users[i].AuthToken.Token] = &users[i]
	}

	return &PersistentInMemoryFileStore{
		fileInteractor: fileInteractor,
		usernameToUser: usernameToUser,
		tokenToUser:    tokenToUser,
		users:          users,
	}, nil
}

func (p *PersistentInMemoryFileStore) CreateUser(username, storedPass string, token entities.Token) (entities.StoredUser, error) {
	newUser := entities.StoredUser{
		Id:         generateId(),
		Username:   username,
		StoredPass: storedPass,
		AuthToken:  token,
	}

	p.users = append(p.users, newUser)
	newUserPtr := &p.users[len(p.users)-1]

	p.usernameToUser[username] = newUserPtr
	p.tokenToUser[token.Token] = newUserPtr

	err := p.fileInteractor.WriteUser(newUser)
	if err != nil {
		return entities.StoredUser{}, errors.New(fmt.Sprintf("Got an error while writing to a file interactor: %v", err))
	}

	return newUser, nil // return a copy, so the caller isn't able to change the user directly
}

func generateId() string {
	// this actually never returns an error
	id, _ := uuid.NewUUID()
	return id.String()
}

func (p *PersistentInMemoryFileStore) FindUser(username string) (entities.StoredUser, error) {
	user, ok := p.usernameToUser[username]
	if !ok {
		return entities.StoredUser{}, auth_store_contract.UserNotFoundErr
	}
	return *user, nil
}
func (p *PersistentInMemoryFileStore) FindUserFromToken(token string) (entities.StoredUser, error) {
	user, ok := p.tokenToUser[token]
	if !ok {
		return entities.StoredUser{}, token_store_contract.TokenNotFoundErr
	}
	return *user, nil
}

func (p *PersistentInMemoryFileStore) UserExists(username string) bool {
	_, exists := p.usernameToUser[username]
	return exists
}

type InMemoryAuthStore struct {
	users []entities.StoredUser
}

func (i *InMemoryAuthStore) UserExists(username string) bool {
	for _, user := range i.users {
		if user.Username == username {
			return true
		}
	}
	return false
}

func (i *InMemoryAuthStore) CreateUser(username string, password string, token entities.Token) {
	i.users = append(i.users, entities.StoredUser{Username: username, StoredPass: password, AuthToken: token})
}

func (i *InMemoryAuthStore) FindUser(username string) (entities.StoredUser, error) {
	for _, user := range i.users {
		if user.Username == username {
			return user, nil
		}
	}
	return entities.StoredUser{}, auth_store_contract.UserNotFoundErr
}

func (i *InMemoryAuthStore) FindUserFromToken(token string) (entities.StoredUser, error) {
	for _, user := range i.users {
		if user.AuthToken.Token == token {
			return user, nil
		}
	}
	return entities.StoredUser{}, token_store_contract.TokenNotFoundErr
}