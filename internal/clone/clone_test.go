package clone

import "testing"

func TestBytesCopies(t *testing.T) {
	orig := []byte{1, 2, 3, 4, 5}
	cp := Bytes(orig)

	// Must be a distinct backing array: mutating the source cannot affect the copy.
	for i := range orig {
		orig[i] = 0
	}
	for i, want := range []byte{1, 2, 3, 4, 5} {
		if cp[i] != want {
			t.Fatalf("Bytes aliased source: cp[%d]=%d, want %d (no defensive copy)", i, cp[i], want)
		}
	}
}

func TestBytesNil(t *testing.T) {
	if got := Bytes(nil); got != nil {
		t.Fatalf("Bytes(nil) = %v, want nil", got)
	}
}

func TestBytesEmpty(t *testing.T) {
	got := Bytes([]byte{})
	if got == nil || len(got) != 0 {
		t.Fatalf("Bytes([]byte{}) = %v, want empty non-nil", got)
	}
}

func TestBytesNonNil(t *testing.T) {
	if got := BytesNonNil(nil); len(got) != 0 {
		t.Fatalf("BytesNonNil(nil) = %v, want empty", got)
	}
	orig := []byte{9, 8, 7}
	cp := BytesNonNil(orig)
	orig[0] = 0
	if cp[0] != 9 {
		t.Fatalf("BytesNonNil aliased source: cp[0]=%d, want 9", cp[0])
	}
}

func TestUintsCopies(t *testing.T) {
	orig := []uint8{1, 2, 3}
	cp := Uints(orig)
	for i := range orig {
		orig[i] = 0
	}
	if cp[0] != 1 || cp[2] != 3 {
		t.Fatalf("Uints aliased source: %v", cp)
	}
}
