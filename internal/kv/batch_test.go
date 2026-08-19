package kv

import (
	"testing"
	"time"
)

func TestApplyBatchAndVerify(t *testing.T) {
	s, err := Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	result, err := s.ApplyBatch([]BatchItem{{Key: "a", Value: []byte("1")}, {Key: "b", Value: []byte("2"), TTL: time.Minute}})
	if err != nil || result.Applied != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	check, err := s.Verify()
	if err != nil || check.Live != 2 {
		t.Fatalf("check=%+v err=%v", check, err)
	}
}
