package helpers

import (
	"math/rand"
	"sync"
	"time"
)

func GenerateRandomTestId() string {
	return GenerateRandomString()
}

var (
	r      = rand.New(rand.NewSource(time.Now().UnixNano()))
	rMutex = &sync.Mutex{}
)

func GenerateRandomString() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, 10)

	rMutex.Lock()
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	rMutex.Unlock()
	return string(b)
}
