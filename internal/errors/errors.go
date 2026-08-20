// Package errors defines the sentinel errors returned by the logkv store.
package errors

import "errors"

// ErrNotFound is returned when a key does not exist, has been deleted, or has
// expired.
var ErrNotFound = errors.New("logkv: key not found")

// ErrClosed is returned when an operation is attempted on a closed store.
var ErrClosed = errors.New("logkv: store closed")

// ErrInvalidKey is returned when a key is empty or otherwise invalid.
var ErrInvalidKey = errors.New("logkv: invalid key")

// ErrTTLExpired is returned by Get when the requested key has expired.
var ErrTTLExpired = errors.New("logkv: key expired")

// ErrAlreadyExists is returned when an atomic create collides with an existing
// live key.
var ErrAlreadyExists = errors.New("logkv: key already exists")

// ErrInvalidRange is returned when a range scan's start key sorts after its end
// key, which would otherwise silently yield an empty result.
var ErrInvalidRange = errors.New("logkv: invalid range")
