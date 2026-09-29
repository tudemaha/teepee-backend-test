package password

import "golang.org/x/crypto/bcrypt"

// Hash generates a bcrypt hash from a plain text password.
func Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check compares a plain text password against a bcrypt hash.
func Check(plain, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
