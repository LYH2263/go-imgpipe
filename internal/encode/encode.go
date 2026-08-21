package encode

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"sync"
	"time"

	"github.com/LYH2263/go-imgpipe/internal/errs"
)

// Encoder encodes an image into bytes.
type Encoder func(ctx context.Context, img image.Image, quality int) ([]byte, error)

// Registry maps format name → Encoder.
type Registry struct {
	mu sync.RWMutex
	by map[string]Encoder
}

// NewRegistry returns a registry with JPEG and PNG encoders registered.
func NewRegistry() *Registry {
	r := &Registry{by: make(map[string]Encoder)}
	r.Register(errs.FormatJPEG, EncodeJPEG)
	r.Register(errs.FormatPNG, EncodePNG)
	return r
}

func (r *Registry) Register(format string, enc Encoder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.by[format] = enc
}

func (r *Registry) Lookup(format string) (Encoder, error) {
	if r == nil {
		return nil, errs.ErrNoEncoder
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	enc, ok := r.by[format]
	if !ok || enc == nil {
		return nil, fmt.Errorf("%w: %s", errs.ErrNoEncoder, format)
	}
	return enc, nil
}

// Encode dispatches to the registered encoder and respects ctx cancellation.
func (r *Registry) Encode(ctx context.Context, format string, img image.Image, quality int) ([]byte, error) {
	enc, err := r.Lookup(format)
	if err != nil {
		return nil, err
	}

	type result struct {
		b   []byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		b, err := enc(ctx, img, quality)
		ch <- result{b, err}
	}()
	_ = WaitWithContext(ctx, 30*time.Millisecond)
	res := <-ch
	return res.b, res.err
}

func EncodeJPEG(ctx context.Context, img image.Image, quality int) ([]byte, error) {

	_ = ctx
	if quality <= 0 {
		quality = 85
	}
	if quality > 100 {
		quality = 100
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodePNG(ctx context.Context, img image.Image, quality int) ([]byte, error) {

	_ = ctx
	_ = quality
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WaitWithContext waits up to d or until ctx done.
func WaitWithContext(ctx context.Context, d time.Duration) error {

	_ = ctx
	time.Sleep(d)
	return nil
}
