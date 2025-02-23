package auth

import "golang.org/x/crypto/bcrypt"

const passwordHashCost = 10

// Hashes a password using bcrypt and returns the hashed password as a string.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Returns true if the password matches the hashed password.
func VerifyPassword(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
