// Package kv provides access to Spin key-value stores.
package kv

import (
	"fmt"
	"iter"

	keyvalue "github.com/spinframework/spin-go-sdk/v3/imports/spin_key_value_3_0_0_key_value"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

// Store represents a connection to a key-value store.
type Store struct {
	store *keyvalue.Store
}

// Open opens the store with the specified label.
func Open(label string) (*Store, error) {
	result := keyvalue.StoreOpen(label)
	if result.IsErr() {
		return nil, errorVariantToError(result.Err())
	}

	return &Store{
		store: result.Ok(),
	}, nil
}

// OpenDefault opens the default store.
//
// This is equivalent to Open("default").
func OpenDefault() (*Store, error) {
	return Open("default")
}

// Set sets the key/value pair in the store.
func (s *Store) Set(key string, value []byte) error {
	result := s.store.Set(key, value)
	if result.IsErr() {
		return errorVariantToError(result.Err())
	}

	return nil
}

// Get returns the value of the provided key from the store.
func (s *Store) Get(key string) ([]byte, error) {
	result := s.store.Get(key)
	if result.IsErr() {
		return nil, errorVariantToError(result.Err())
	}

	value := result.Ok()
	if value.IsNone() {
		return []byte(""), nil
	}

	return value.Some(), nil
}

// Delete removes the given key/value from the store.
func (s *Store) Delete(key string) error {
	result := s.store.Delete(key)
	if result.IsErr() {
		return errorVariantToError(result.Err())
	}

	return nil
}

// Exists checks if a given key exists in the store.
func (s *Store) Exists(key string) (bool, error) {
	result := s.store.Exists(key)
	if result.IsErr() {
		return false, errorVariantToError(result.Err())
	}

	return result.Ok(), nil
}

// Keys allows iterating over keys from a key-value store. Use All to iterate
// over the keys as they are read from the stream, then call Err to check for
// any errors. Close must be called when done to release resources.
type Keys struct {
	stream *wittypes.StreamReader[string]
	future *wittypes.FutureReader[wittypes.Result[wittypes.Unit, keyvalue.Error]]
	err    error
}

// GetKeys returns a Keys iterator for all keys in the store.
//
// The caller must call Close on the returned Keys when done.
func (s *Store) GetKeys() *Keys {
	stream, future := s.store.GetKeys()
	return &Keys{
		stream: stream,
		future: future,
	}
}

// All returns an iterator that yields keys as they are read from the
// underlying stream. After iteration completes, call Err to check for errors.
func (k *Keys) All() iter.Seq[string] {
	return func(yield func(string) bool) {
		buf := make([]string, 64)
		for {
			n := k.stream.Read(buf)
			for i := range n {
				if !yield(buf[i]) {
					return
				}
			}
			if k.stream.WriterDropped() {
				break
			}
		}
		result := k.future.Read()
		if result.IsErr() {
			k.err = errorVariantToError(result.Err())
		}
	}
}

// Err returns any error encountered during iteration. It must be called after
// iteration completes.
func (k *Keys) Err() error {
	return k.err
}

// Close releases resources associated with the key iterator.
func (k *Keys) Close() {
	k.stream.Drop()
}

func errorVariantToError(code keyvalue.Error) error {
	switch code.Tag() {
	case keyvalue.ErrorAccessDenied:
		return fmt.Errorf("access denied")
	case keyvalue.ErrorNoSuchStore:
		return fmt.Errorf("no such store")
	case keyvalue.ErrorStoreTableFull:
		return fmt.Errorf("store table full")
	case keyvalue.ErrorOther:
		return fmt.Errorf("%v", code.Other())
	default:
		return fmt.Errorf("no error provided by host implementation")
	}
}
