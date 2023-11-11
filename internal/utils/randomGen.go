package utils

import (
	"crypto/rand"
	"fmt"
	"io"
)

func main() {
	for i := 0; i < 3; i++ {
		fmt.Println(EncodeToString(6))
	}
}

func EncodeToString(max int) (string, error) {
	b := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, b, max)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
