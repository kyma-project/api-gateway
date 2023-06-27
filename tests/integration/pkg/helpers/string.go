package helpers

import (
	"math/rand"
	"time"
)

func GenerateRandomTestId() string {
	return GenerateRandomString(8)
}

func GenerateRandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}
