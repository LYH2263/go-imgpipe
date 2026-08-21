package persist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-imgpipe/internal/persist"
)

func TestBug06_RenameFailNoIndex(t *testing.T) {
	dir := t.TempDir()
	st, err := persist.Open(persist.Options{Dir: dir, SyncOnWrite: true})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	key := "frame-rename-fail"
	final := filepath.Join(dir, "blobs", key+".bin")
	if err := os.MkdirAll(final, 0o755); err != nil {
		t.Fatal(err)
	}
	err = st.Put(&persist.Entry{
		Key:    key,
		W:      2,
		H:      2,
		Stride: 8,
		Format: "png",
		Pixels: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Raw:    []byte("raw"),
	})
	if err == nil {
		t.Fatal("expected rename failure")
	}
	if st.Has(key) {
		t.Fatal("memory index updated after rename failure")
	}
}
