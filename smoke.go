package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

// SmokeTest exercises the core store surface end to end without binding a
// network port. It returns nil on success.
func SmokeTest() error {
	dir, err := os.MkdirTemp("", "logkv-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	store, err := kv.Open(filepath.Join(dir, "smoke.db"))
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Put("alpha", []byte("1")); err != nil {
		return err
	}
	if err := store.Put("beta", []byte("2")); err != nil {
		return err
	}
	if err := store.Put("gamma", []byte("3")); err != nil {
		return err
	}

	v, err := store.Get("beta")
	if err != nil {
		return err
	}
	if string(v) != "2" {
		return fmt.Errorf("get beta = %q, want 2", v)
	}

	if err := store.PutWithTTL("temp", []byte("x"), 3*time.Second); err != nil {
		return err
	}
	if _, err := store.Get("temp"); err != nil {
		return err
	}

	rs, err := store.RangeScan("alpha", "gamma")
	if err != nil {
		return err
	}
	if len(rs) != 2 || rs[0].Key != "alpha" || rs[1].Key != "beta" {
		return fmt.Errorf("range = %+v", rs)
	}

	ks, err := store.Keys("a")
	if err != nil {
		return err
	}
	if len(ks) != 1 || ks[0] != "alpha" {
		return fmt.Errorf("keys = %v", ks)
	}

	if err := store.Delete("alpha"); err != nil {
		return err
	}
	if _, err := store.Get("alpha"); err == nil {
		return fmt.Errorf("deleted alpha still found")
	}

	var buf bytes.Buffer
	if err := store.Backup(&buf); err != nil {
		return err
	}
	if err := store.Restore(&buf); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.Export(ctx, &bytes.Buffer{}); err == nil {
		return fmt.Errorf("export ignored cancelled context")
	}

	if err := store.Compact(); err != nil {
		return err
	}

	st, err := store.Stats()
	if err != nil {
		return err
	}
	if st.Count == 0 {
		return fmt.Errorf("stats count 0")
	}

	if err := store.PutWithTTL("ephemeral", []byte("e"), 400*time.Millisecond); err != nil {
		return err
	}
	time.Sleep(900 * time.Millisecond)
	if _, err := store.ExpireNow(); err != nil {
		return err
	}
	if _, err := store.Get("ephemeral"); err == nil {
		return fmt.Errorf("expired key still found")
	}

	return nil
}
