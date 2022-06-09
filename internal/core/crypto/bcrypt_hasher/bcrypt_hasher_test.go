package bcrypt_hasher_test

import (
	"testing"
	"testing/quick"

	"internal/core/crypto/bcrypt_hasher"
	. "internal/test_helpers"
)

func TestBcryptHasher(t *testing.T) {
	t.Run("property based test", func(t *testing.T) {
		hasher := bcrypt_hasher.NewBcryptHasher(5)
		assertion := func(pass string) bool {
			// random string must be considered not hashed
			if hasher.IsHashed(pass) {
				return false
			}
			hashedPass, err := hasher.Hash(pass)
			if err != nil {
				return false
			}
			if !hasher.IsHashed(hashedPass) {
				return false
			}
			if !hasher.Compare(pass, hashedPass) {
				return false
			}
			return true
		}

		if err := quick.Check(assertion, &quick.Config{MaxCount: 150}); err != nil {
			t.Error("failed checks", err)
		}
	})

	t.Run("doesn't ignore errors", func(t *testing.T) {
		hasher := bcrypt_hasher.NewBcryptHasher(10000)
		_, err := hasher.Hash(RandomString())
		AssertSomeError(t, err)
	})
}
