package imgpipe

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCachePutDiskFailureNoMemIndex reproduces the stress-test scenario:
// blobs/<key>.bin is pre-created as a directory so the disk rename fails.
// Cache.Put must return the error AND must not write the key into the
// memory index — otherwise a later Has/Get falsely reports a cache hit.
func TestCachePutDiskFailureNoMemIndex(t *testing.T) {
	dir := t.TempDir()
	p, err := New(Options{CacheDir: dir, MemLimit: 16})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	c := p.Cache()

	key := "deadbeefdeadbeef"
	// Sabotage: make blobs/<key>.bin a directory so os.Rename fails.
	blobPath := filepath.Join(dir, "blobs", key+".bin")
	if err := os.MkdirAll(blobPath, 0o755); err != nil {
		t.Fatalf("mkdir sabotage: %v", err)
	}

	fr := &Frame{W: 1, H: 1, Stride: 1, Format: "RGBA", Pixels: []byte{1, 2, 3, 4}}
	if err := c.Put(key, fr); err == nil {
		t.Fatalf("Cache.Put: expected disk error, got nil")
	}

	// The whole point: no phantom hit in memory or on disk.
	if c.Has(key) {
		t.Fatalf("Cache.Has: phantom hit after disk write failure")
	}
	if _, ok := c.Get(key); ok {
		t.Fatalf("Cache.Get: phantom hit after disk write failure")
	}
	if p.mem.Len() != 0 {
		t.Fatalf("mem index polluted after disk write failure: len=%d", p.mem.Len())
	}
}
