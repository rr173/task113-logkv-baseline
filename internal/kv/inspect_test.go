package kv

import "testing"

func TestInspectNamespaces(t *testing.T) {
	s, err := Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("user/a", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("user/b", []byte("bb")); err != nil {
		t.Fatal(err)
	}
	items, err := s.Inspect("user/")
	if err != nil || len(items) != 2 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	groups, err := s.Namespaces("/")
	if err != nil || len(groups) != 1 || groups[0].Keys != 2 {
		t.Fatalf("groups=%v err=%v", groups, err)
	}
}
