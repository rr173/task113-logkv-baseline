package kv

import "testing"

func TestBug05QueryNormalizesReversedSizeBounds(t *testing.T) {
	store, err := Open(t.TempDir() + "/store.db")
	if err != nil { t.Fatal(err) }
	defer store.Close()
	if err := store.Put("small", []byte("123")); err != nil { t.Fatal(err) }
	if err := store.Put("medium", []byte("123456")); err != nil { t.Fatal(err) }
	items, err := store.Query(Query{MinSize: 8, MaxSize: 2})
	if err != nil { t.Fatal(err) }
	if len(items) != 2 { t.Fatalf("items = %+v, want both normalized matches", items) }
}
