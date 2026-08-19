package errors

import "errors"

func IsMissing(err error) bool       { return errors.Is(err, ErrNotFound) || errors.Is(err, ErrTTLExpired) }
func IsWriteConflict(err error) bool { return errors.Is(err, ErrAlreadyExists) }
