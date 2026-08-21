package persist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-imgpipe/internal/persist"
)

func TestBug09_CacheTempSyncBeforeClose(t *testing.T) {
	dir := t.TempDir()
	st, err := persist.Open(persist.Options{Dir: dir, SyncOnWrite: true})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	key := "sync-before-close"
	pixels := []byte{
		10, 20, 30, 255, 11, 21, 31, 255,
		12, 22, 32, 255, 13, 23, 33, 255,
	}
	raw := []byte("payload-raw-bytes")
	if err := st.Put(&persist.Entry{
		Key: key, W: 2, H: 2, Stride: 8, Format: "png", Pixels: pixels, Raw: raw,
	}); err != nil {
		t.Fatalf("Put failed (Close before Sync?): %v", err)
	}
	path := filepath.Join(dir, "blobs", key+".bin")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 20 {
		t.Fatalf("blob too short: %d", len(got))
	}
	ent, err := st.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(ent.Raw) != string(raw) {
		t.Fatalf("raw=%q want %q", ent.Raw, raw)
	}
}
