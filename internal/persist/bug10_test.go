package persist_test

import (
	"testing"

	"github.com/LYH2263/go-imgpipe/internal/persist"
)

func TestBug10_CloseSyncsJournalBeforeUnmount(t *testing.T) {
	dir := t.TempDir()
	st, err := persist.Open(persist.Options{Dir: dir, SyncOnWrite: true})
	if err != nil {
		t.Fatal(err)
	}
	key := "journal-close-order"
	if err := st.Put(&persist.Entry{
		Key: key, W: 1, H: 1, Stride: 4, Format: "png",
		Pixels: []byte{1, 2, 3, 4}, Raw: []byte("x"),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("Close failed (Sync after Close on journal?): %v", err)
	}
	st2, err := persist.Open(persist.Options{Dir: dir, SyncOnWrite: true})
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	if !st2.Has(key) {
		t.Fatal("journal entry lost after Close without prior Sync")
	}
}
