package helpers

import (
	"hash/maphash"
	"math/rand"
)

func GenerateRandomTestId() string {
	return GenerateRandomString(8)
}

func GenerateRandomString(length int) string {
	rand.NewSource(int64(new(maphash.Hash).Sum64()))

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
