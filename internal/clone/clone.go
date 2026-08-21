package clone

// Bytes returns a defensive copy of b. Nil input yields nil.
//
// The copy is essential: callers use it to sever aliasing between a caller-owned
// input buffer and pipeline-owned frames/cache entries. Returning b unchanged
// would let an external mutation (e.g. zeroing the upload buffer after Run)
// bleed into Result.Frame.Raw and the in-memory cache entry through a shared
// backing array.
func Bytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// BytesNonNil always returns a non-nil slice (possibly empty).
func BytesNonNil(b []byte) []byte {
	if b == nil {
		return []byte{}
	}
	return Bytes(b)
}

// Strings copies a string slice.
func Strings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

// Uints copies a uint8 slice (alias of Bytes for clarity at call sites).
func Uints(b []uint8) []uint8 {
	return Bytes(b)
}
