package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// returns a hex encoded random byte string, double the length of s
func GenerateToken(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return hex.EncodeToString(b), err
}
