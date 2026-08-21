package cachekey_test

import (
        "testing"

        "github.com/LYH2263/go-imgpipe/internal/cachekey"
)

func TestComputeStable(t *testing.T) {
        raw := []byte{1, 2, 3, 4}
        ts := []cachekey.Spec{{Kind: "scale", Width: 10, Height: 10}}
        enc := cachekey.EncodeSpec{Format: "jpeg", Quality: 80}
        a := cachekey.Compute(raw, ts, enc, true)
        b := cachekey.Compute(raw, ts, enc, true)
        if a != b || len(a) != 64 {
                t.Fatalf("%s vs %s", a, b)
        }
        c := cachekey.Compute(raw, ts, enc, false)
        if a == c {
                t.Fatal("strip flag should change key")
        }
}
