package bcrypt_hasher

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	hashCost int
}

func NewBcryptHasher(hashCost int) *BcryptHasher {
	return &BcryptHasher{hashCost: hashCost}
}

func (b BcryptHasher) IsHashed(pass string) bool {
	_, err := bcrypt.Cost([]byte(pass))
	return err == nil
}

func (b BcryptHasher) Hash(pass string) (string, error) {
	hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(pass), b.hashCost)
	if err != nil {
		return "", errors.New(fmt.Sprintf("hashing failed: %v", err))
	}
	return string(hashedPassBytes), nil
}

func (b BcryptHasher) Compare(pass, hashedPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass))
	return err == nil
}
