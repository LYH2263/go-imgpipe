package imgpipe_test

import (
        "bytes"
        "context"
        "image"
        "image/color"
        "image/jpeg"
        "image/png"
        "os"
        "path/filepath"
        "testing"
        "time"

        "github.com/LYH2263/go-imgpipe"
)

func makeJPEG(t *testing.T, w, h int) []byte {
        t.Helper()
        img := image.NewRGBA(image.Rect(0, 0, w, h))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 80, A: 255})
                }
        }
        var buf bytes.Buffer
        if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
                t.Fatal(err)
        }
        return buf.Bytes()
}

func makePNG(t *testing.T, w, h int) []byte {
        t.Helper()
        img := image.NewRGBA(image.Rect(0, 0, w, h))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        img.Set(x, y, color.RGBA{R: 10, G: uint8(x), B: uint8(y), A: 255})
                }
        }
        var buf bytes.Buffer
        if err := png.Encode(&buf, img); err != nil {
                t.Fatal(err)
        }
        return buf.Bytes()
}

func TestRunScaleJPEG(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        raw := makeJPEG(t, 64, 48)
        res, err := p.Run(imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{
                        Kind: imgpipe.TransformScale, Width: 32, Height: 24, Filter: imgpipe.FilterBilinear,
                }},
                Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 85},
        })
        if err != nil {
                t.Fatal(err)
        }
        if res.Width != 32 || res.Height != 24 {
                t.Fatalf("size %dx%d", res.Width, res.Height)
        }
        if len(res.Bytes) == 0 || res.CacheKey == "" {
                t.Fatal("empty output or key")
        }
}

func TestCacheHitAndIsolation(t *testing.T) {
        dir := t.TempDir()
        p, err := imgpipe.New(imgpipe.Options{CacheDir: dir})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        raw := makePNG(t, 40, 40)
        job := imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{
                        Kind: imgpipe.TransformScale, Width: 20, Height: 20, Filter: imgpipe.FilterNearest,
                }},
                Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodePNG},
        }
        r1, err := p.Run(job)
        if err != nil {
                t.Fatal(err)
        }
        if r1.CacheHit {
                t.Fatal("first should miss")
        }
        // mutate caller raw — must not affect cache
        raw[10] ^= 0xff
        r2, err := p.Run(imgpipe.Job{
                Raw: append([]byte(nil), makePNG(t, 40, 40)...),
                Transforms: job.Transforms,
                Encode:     job.Encode,
        })
        // use original bytes again
        raw2 := makePNG(t, 40, 40)
        r2, err = p.Run(imgpipe.Job{Raw: raw2, Transforms: job.Transforms, Encode: job.Encode})
        if err != nil {
                t.Fatal(err)
        }
        if !r2.CacheHit {
                t.Fatal("expected cache hit")
        }
        // Cache.Get must return independent pixel slice
        fr, ok := p.Cache().Get(r1.CacheKey)
        if !ok {
                t.Fatal("missing cache")
        }
        fr.Pixels[0] ^= 0xff
        fr2, ok := p.Cache().Get(r1.CacheKey)
        if !ok {
                t.Fatal("missing")
        }
        if fr2.Pixels[0] == fr.Pixels[0] && len(fr2.Pixels) > 0 {
                // if both flipped same value coincidence; check they differ from mutated expectation
                // mutated fr should not equal second get's original — re-get and compare to fr
        }
        if fr2.Pixels[0] == fr.Pixels[0] {
                // same means shared buffer (both show flipped) — fail
                // Actually after flip fr.Pixels[0] is flipped; if shared, fr2 also flipped.
                // Re-fetch third time from cache internal — Get always copies so fr2 should be original.
                // If fr and fr2 share, they'd be equal (both flipped). So equal => BUG.
                t.Fatal("cache Get appears to share pixels with caller")
        }
}

func TestCropAndRotate(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        raw := makeJPEG(t, 50, 50)
        res, err := p.Run(imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{
                        {Kind: imgpipe.TransformCrop, X: 5, Y: 5, Width: 30, Height: 30},
                        {Kind: imgpipe.TransformRotate, Degrees: 90},
                },
                Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodePNG},
        })
        if err != nil {
                t.Fatal(err)
        }
        if res.Width != 30 || res.Height != 30 {
                t.Fatalf("got %dx%d", res.Width, res.Height)
        }
}

func TestClosedRun(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        if err := p.Close(); err != nil {
                t.Fatal(err)
        }
        _, err = p.Run(imgpipe.Job{Raw: makeJPEG(t, 8, 8)})
        if err != imgpipe.ErrClosed {
                t.Fatalf("got %v", err)
        }
}

func TestUnsupportedFormatWrapped(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        _, err = p.Run(imgpipe.Job{Raw: []byte("not-an-image")})
        if err == nil {
                t.Fatal("expected error")
        }
}

func TestContextCancel(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        ctx, cancel := context.WithCancel(context.Background())
        cancel()
        _, err = p.RunContext(ctx, imgpipe.Job{Raw: makeJPEG(t, 16, 16)})
        if err == nil {
                t.Fatal("expected cancel")
        }
}

func TestDiskCacheRoundTrip(t *testing.T) {
        dir := t.TempDir()
        p, err := imgpipe.New(imgpipe.Options{CacheDir: dir})
        if err != nil {
                t.Fatal(err)
        }
        raw := makeJPEG(t, 24, 24)
        job := imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{Kind: imgpipe.TransformScale, Width: 12, Height: 12}},
                Encode:     imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 70},
        }
        r1, err := p.Run(job)
        if err != nil {
                t.Fatal(err)
        }
        if err := p.Close(); err != nil {
                t.Fatal(err)
        }
        p2, err := imgpipe.New(imgpipe.Options{CacheDir: dir})
        if err != nil {
                t.Fatal(err)
        }
        defer p2.Close()
        // memory empty; disk should still have after Put — but mem was cleared on close.
        // New pipeline replays journal.
        if !p2.Cache().Has(r1.CacheKey) {
                // Has checks mem then disk
                t.Log("keys on disk", filepath.Join(dir, "blobs"))
                ents, _ := os.ReadDir(filepath.Join(dir, "blobs"))
                t.Logf("blob count %d journal", len(ents))
                if !p2.Cache().Has(r1.CacheKey) {
                        t.Fatal("expected disk index after reopen")
                }
        }
        _ = time.Second
}

func TestInputCopyProtectsCache(t *testing.T) {
        p, err := imgpipe.New(imgpipe.Options{})
        if err != nil {
                t.Fatal(err)
        }
        defer p.Close()
        raw := makeJPEG(t, 20, 20)
        job := imgpipe.Job{Raw: raw, Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 80}}
        r1, err := p.Run(job)
        if err != nil {
                t.Fatal(err)
        }
        // mutate input after run
        for i := range raw {
                raw[i] = 0
        }
        fr, ok := p.Cache().Get(r1.CacheKey)
        if !ok || fr == nil || len(fr.Raw) == 0 {
                t.Fatal("cache missing")
        }
        // cached raw must not be all zeros
        allZero := true
        for _, b := range fr.Raw {
                if b != 0 {
                        allZero = false
                        break
                }
        }
        if allZero {
                t.Fatal("cache raw polluted by caller mutation")
        }
}
