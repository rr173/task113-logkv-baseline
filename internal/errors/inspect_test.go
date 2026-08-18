package errors

import "testing"

func TestClassify(t *testing.T) {
	if !IsMissing(ErrTTLExpired) || !IsWriteConflict(ErrAlreadyExists) {
		t.Fatal("sentinel classification failed")
	}
}
