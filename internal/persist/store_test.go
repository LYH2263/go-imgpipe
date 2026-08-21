package persist

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPutRenameFailureNoIndex reproduces the stress-test scenario:
// blobs/<key>.bin is pre-created as a directory, so os.Rename fails.
// The memory index must NOT record the key — otherwise Has returns a
// phantom hit for a blob that never landed on disk.
func TestPutRenameFailureNoIndex(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(Options{Dir: dir, JournalName: "cache.journal"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	key := "deadbeef"
	// Sabotage the blob path: make blobs/<key>.bin a directory so the
	// final rename target cannot be created as a regular file.
	blobDir := filepath.Join(dir, "blobs", key+".bin")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		t.Fatalf("mkdir sabotage: %v", err)
	}

	e := &Entry{Key: key, W: 2, H: 2, Stride: 2, Format: "RGBA", Pixels: []byte{0, 0, 0, 0}}
	if err := s.Put(e); err == nil {
		t.Fatalf("Put: expected rename error, got nil")
	}

	if s.Has(key) {
		t.Fatalf("Has: key indexed after failed rename — phantom hit")
	}
	if _, ok := s.index[key]; ok {
		t.Fatalf("internal index records key after failed rename")
	}
}

// TestPutRenameFailureNoJournal verifies no journal "put" record is
// appended when rename fails, so a reopened store stays consistent.
func TestPutRenameFailureNoJournal(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(Options{Dir: dir, JournalName: "cache.journal"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	key := "cafebabe"
	_ = os.MkdirAll(filepath.Join(dir, "blobs", key+".bin"), 0o755)
	e := &Entry{Key: key, W: 1, H: 1, Stride: 1, Format: "RGB", Pixels: []byte{1, 2, 3}}
	_ = s.Put(e) // expected to fail
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Reopen — the failed put must not have been journaled.
	s2, err := Open(Options{Dir: dir, JournalName: "cache.journal"})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	if s2.Has(key) {
		t.Fatalf("reopened store has phantom key from failed put")
	}
}
