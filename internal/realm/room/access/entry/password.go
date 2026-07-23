package entry

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a non-empty room password with bcrypt.
func HashPassword(password string, cost int) (string, error) {
	if password == "" {
		return "", ErrInvalidPassword
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// passwordMatches reports whether plaintext matches an optional bcrypt hash.
func passwordMatches(hash *string, plaintext string) bool {
	if hash == nil || plaintext == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(*hash), []byte(plaintext)) == nil
}
