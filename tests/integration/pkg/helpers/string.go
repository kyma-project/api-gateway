package helpers

import (
	"math/rand"
	"sync"
	"time"
)

func GenerateRandomTestId() string {
	return GenerateRandomString(8)
}

var (
	r      = rand.New(rand.NewSource(time.Now().UnixNano()))
	rMutex = &sync.Mutex{}
)

func GenerateRandomString(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)

	rMutex.Lock()
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	rMutex.Unlock()
	return string(b)
}
