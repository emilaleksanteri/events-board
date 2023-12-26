package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"
)

func makeMac(token string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(token))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func CheckMac(token, mac string) bool {
	key := []byte(os.Getenv("SESSION_SECRET"))
	return hmac.Equal([]byte(mac), []byte(makeMac(token, key)))
}

func MakeToken(tokenToHash string) string {
	secret := os.Getenv("SESSION_SECRET")
	return makeMac(tokenToHash, []byte(secret))
}
