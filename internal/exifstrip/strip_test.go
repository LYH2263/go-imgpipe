package exifstrip_test

import (
        "testing"

        "github.com/LYH2263/go-imgpipe/internal/exifstrip"
)

func TestStripNonJPEG(t *testing.T) {
        in := []byte("hello")
        out := exifstrip.StripJPEG(in)
        if string(out) != "hello" {
                t.Fatal(out)
        }
}
