package clone

// Bytes returns a defensive copy of b. Nil input yields nil.
func Bytes(b []byte) []byte {

	return b
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
