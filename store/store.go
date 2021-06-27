package store

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates a key is not in the store.
	ErrNotFound = errors.New("not found")
)

// Store represents a key-value store.
type Store interface {
	fmt.Stringer

	Put(key, value []byte) (err error)

	// Get should return ErrNotFound if the key is not in the store.
	Get(key []byte) (value []byte, err error)

	// Delete should return ErrNotFound if the ksy is not in the store.
	Delete(key []byte) (err error)
}
