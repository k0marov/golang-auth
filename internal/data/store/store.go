package store

import (
	"fmt"
	"sync"

	"internal/data/models"
	"internal/domain/auth_store_contract"
	"internal/domain/entities"
	"internal/domain/token_store_contract"
)

type DBFileInteractor interface {
	ReadUsers() ([]models.UserModel, error)
	WriteUser(models.UserModel) error
}

// An in-memory database is used here for 2 reasons:
// 1. This database is accessed on nearly every request (see TokenAuthMiddleware), so speed is needed
// 2. The schema is quite light on memory - 50 MB of RAM is enough to hold 100 000+ users
type PersistentInMemoryFileStore struct {
	fileInteractor DBFileInteractor

	usernameToUser map[string]*models.UserModel
	tokenToUser    map[string]*models.UserModel
	users          []models.UserModel

	biggestId int

	mu sync.Mutex
}

func NewPersistentInMemoryFileStore(fileInteractor DBFileInteractor) (*PersistentInMemoryFileStore, error) {
	users, err := fileInteractor.ReadUsers()
	if err != nil {
		return nil, fmt.Errorf("got an error while reading users from file interactor: %w", err)
	}

	biggestId := 0
	usernameToUser := make(map[string]*models.UserModel)
	tokenToUser := make(map[string]*models.UserModel)

	for i := range users {
		user := &users[i]
		if user.Id > biggestId {
			biggestId = user.Id
		}
		usernameToUser[user.Username] = user
		tokenToUser[user.AuthToken.Token] = user
	}

	return &PersistentInMemoryFileStore{
		fileInteractor: fileInteractor,
		usernameToUser: usernameToUser,
		tokenToUser:    tokenToUser,
		users:          users,
		biggestId:      biggestId,
	}, nil
}

func (p *PersistentInMemoryFileStore) CreateUser(username, storedPass string, token entities.Token) (models.UserModel, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	newUser := models.UserModel{
		Id:         p.biggestId + 1,
		Username:   username,
		StoredPass: storedPass,
		AuthToken:  token,
	}

	err := p.fileInteractor.WriteUser(newUser)
	if err != nil {
		return models.UserModel{}, fmt.Errorf("got an error while writing to a file interactor: %w", err)
	}

	p.users = append(p.users, newUser)
	newUserPtr := &p.users[len(p.users)-1]

	p.usernameToUser[username] = newUserPtr
	p.tokenToUser[token.Token] = newUserPtr

	p.biggestId++

	return newUser, nil // return a copy, so the caller is not able to change the user directly
}

func (p *PersistentInMemoryFileStore) FindUser(username string) (models.UserModel, error) {
	user, ok := p.usernameToUser[username]
	if !ok {
		return models.UserModel{}, auth_store_contract.UserNotFoundErr
	}
	return *user, nil
}
func (p *PersistentInMemoryFileStore) FindUserFromToken(token string) (models.UserModel, error) {
	user, ok := p.tokenToUser[token]
	if !ok {
		return models.UserModel{}, token_store_contract.TokenNotFoundErr
	}
	return *user, nil
}

func (p *PersistentInMemoryFileStore) UserExists(username string) bool {
	_, exists := p.usernameToUser[username]
	return exists
}
