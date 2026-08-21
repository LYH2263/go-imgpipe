package cachekey

import (
        "crypto/sha256"
        "encoding/hex"
        "fmt"
        "hash"
        "io"
        "strconv"
)

// Spec is a transform description for hashing (no dependency on parent types).
type Spec struct {
        Kind    string
        Width   int
        Height  int
        Filter  string
        X       int
        Y       int
        Overlay []byte
        Opacity float64
        Degrees int
        FlipH   bool
        FlipV   bool
}

// EncodeSpec is encode options for hashing.
type EncodeSpec struct {
        Format  string
        Quality int
}

// Compute builds a stable content-addressed cache key.
func Compute(raw []byte, transforms []Spec, enc EncodeSpec, strip bool) string {
        h := sha256.New()
        _, _ = h.Write([]byte("imgpipe/v1\n"))
        writeBool(h, strip)
        writeBytes(h, raw)
        writeTransforms(h, transforms)
        writeEncode(h, enc)
        sum := h.Sum(nil)
        return hex.EncodeToString(sum[:])
}

// Short returns the first n hex chars of Compute (n clamped to [8,64]).
func Short(raw []byte, transforms []Spec, enc EncodeSpec, strip bool, n int) string {
        full := Compute(raw, transforms, enc, strip)
        if n < 8 {
                n = 8
        }
        if n > len(full) {
                n = len(full)
        }
        return full[:n]
}

func writeBool(h hash.Hash, v bool) {
        if v {
                _, _ = h.Write([]byte{1})
        } else {
                _, _ = h.Write([]byte{0})
        }
}

func writeBytes(h hash.Hash, b []byte) {
        writeInt(h, len(b))
        _, _ = h.Write(b)
}

func writeInt(h hash.Hash, n int) {
        var buf [8]byte
        u := uint64(n)
        for i := 0; i < 8; i++ {
                buf[i] = byte(u >> (8 * i))
        }
        _, _ = h.Write(buf[:])
}

func writeTransforms(h hash.Hash, ts []Spec) {
        writeInt(h, len(ts))
        for _, t := range ts {
                _, _ = io.WriteString(h, t.Kind)
                writeInt(h, t.Width)
                writeInt(h, t.Height)
                writeInt(h, t.X)
                writeInt(h, t.Y)
                writeInt(h, t.Degrees)
                _, _ = io.WriteString(h, t.Filter)
                writeBool(h, t.FlipH)
                writeBool(h, t.FlipV)
                writeInt(h, int(t.Opacity*1000))
                writeBytes(h, t.Overlay)
        }
}

func writeEncode(h hash.Hash, e EncodeSpec) {
        _, _ = io.WriteString(h, e.Format)
        writeInt(h, e.Quality)
}

// FormatHuman returns a debug string (not used as cache key).
func FormatHuman(rawLen int, nTransforms int, format string, quality int) string {
        return fmt.Sprintf("raw=%d transforms=%d encode=%s/%s",
                rawLen, nTransforms, format, strconv.Itoa(quality))
}
