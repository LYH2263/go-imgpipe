package imgpipe

import "testing"

// TestCacheGetIsolatesPixels reproduces the management-page "highlight first pixel"
// scenario: a caller mutates the Frame returned by Cache.Get, then the same process
// Get()s the same key again. The cached entry must NOT be dirty — i.e. Get must
// return a deep copy of entry.Pixels (and Raw), not an alias into the internal slice.
func TestCacheGetIsolatesPixels(t *testing.T) {
	p, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	cache := p.Cache()

	orig := &Frame{
		Pixels: []uint8{1, 2, 3, 4, 5, 6, 7, 8},
		W:      1, H: 1, Stride: 4, Format: string(EncodePNG),
	}
	if err := cache.Put("k", orig); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Get #1: highlight first pixel through the returned Frame.
	got, ok := cache.Get("k")
	if !ok {
		t.Fatal("Get #1: miss")
	}
	if got.Pixels[0] != 1 {
		t.Fatalf("Get #1: pixels[0] = %d, want 1", got.Pixels[0])
	}
	got.Pixels[0] = 0xFE // simulate admin-page highlight
	got.Pixels[1] = 0xFE

	// Get #2: must observe the original bytes, not the mutation.
	again, ok := cache.Get("k")
	if !ok {
		t.Fatal("Get #2: miss")
	}
	if again.Pixels[0] != 1 || again.Pixels[1] != 2 {
		t.Fatalf("Get #2: entry was mutated by Get #1; pixels = %v, want [1 2 ...]", again.Pixels)
	}

	// And mutating #2 must not affect #3.
	again.Pixels[0] = 0x42
	third, _ := cache.Get("k")
	if third.Pixels[0] != 1 {
		t.Fatalf("Get #3: entry was mutated by Get #2; pixels[0] = %d, want 1", third.Pixels[0])
	}
}

// TestCacheGetIsolatesRaw guards the Raw byte slice the same way.
func TestCacheGetIsolatesRaw(t *testing.T) {
	p, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	cache := p.Cache()
	orig := &Frame{
		Pixels: []uint8{0, 0, 0, 0},
		Raw:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
		W:      1, H: 1, Stride: 4,
	}
	if err := cache.Put("k", orig); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, _ := cache.Get("k")
	got.Raw[0] = 0x00

	again, _ := cache.Get("k")
	if again.Raw[0] != 0xDE {
		t.Fatalf("Get #2: Raw was mutated by Get #1; raw = % x, want de ad be ef", again.Raw)
	}
}
