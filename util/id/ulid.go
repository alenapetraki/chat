package id

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/pkg/errors"
)

var mu sync.Mutex
var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

// есть ли разница, работать с ULID или строкой?

func MustNewULID() string {
	result, err := NewULID()
	if err != nil {
		panic(err)
	}
	return result
}

func NewULID() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	result, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		return "", errors.Wrap(err, "could not generate a new ULID")
	}
	return result.String(), nil
}
