package clone

// StringMap shallow-copies a map[string]string.
func StringMap(in map[string]string) map[string]string {
        if in == nil {
                return nil
        }
        out := make(map[string]string, len(in))
        for k, v := range in {
                out[k] = v
        }
        return out
}

// ByteMap copies map[string][]byte with deep byte copies.
func ByteMap(in map[string][]byte) map[string][]byte {
        if in == nil {
                return nil
        }
        out := make(map[string][]byte, len(in))
        for k, v := range in {
                out[k] = Bytes(v)
        }
        return out
}
