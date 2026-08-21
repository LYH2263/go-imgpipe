package imgpipe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LYH2263/go-imgpipe/internal/cachekey"
	"github.com/LYH2263/go-imgpipe/internal/clone"
	"github.com/LYH2263/go-imgpipe/internal/decode"
	"github.com/LYH2263/go-imgpipe/internal/encode"
	"github.com/LYH2263/go-imgpipe/internal/exifstrip"
	"github.com/LYH2263/go-imgpipe/internal/metrics"
	"github.com/LYH2263/go-imgpipe/internal/persist"
	"github.com/LYH2263/go-imgpipe/internal/xform"
)

// Pipeline runs decode → transform → encode jobs with optional caching.
type Pipeline struct {
	opts     Options
	syncOn   bool
	mu       sync.Mutex
	closed   int32
	mem      *memIndex
	disk     *persist.Store
	encoders *encode.Registry
	metrics  *metrics.Collector
	cache    *Cache
}

// New constructs a Pipeline. CacheDir may be empty for memory-only mode.
func New(opts Options) (*Pipeline, error) {
	o := opts.withDefaults()
	syncOn := true
	p := &Pipeline{
		opts:     o,
		syncOn:   syncOn,
		mem:      newMemIndex(o.MemLimit),
		encoders: encode.NewRegistry(),
		metrics:  metrics.New(),
	}
	p.cache = &Cache{pipe: p}
	if o.CacheDir != "" {
		st, err := persist.Open(persist.Options{
			Dir:         o.CacheDir,
			JournalName: o.JournalName,
			SyncOnWrite: syncOn,
			LimitBytes:  o.DiskLimitBytes,
		})
		if err != nil {
			return nil, err
		}
		p.disk = st
	}
	return p, nil
}

// Cache returns the content cache handle.
func (p *Pipeline) Cache() *Cache {
	if p == nil {
		return nil
	}
	return p.cache
}

// Stats returns a snapshot of counters.
func (p *Pipeline) Stats() Stats {
	if p == nil {
		return Stats{Closed: true}
	}
	s := p.metrics.Snapshot()
	return Stats{
		Runs:       s.Runs,
		CacheHits:  s.CacheHits,
		CacheMiss:  s.CacheMiss,
		DecodeFail: s.DecodeFail,
		EncodeFail: s.EncodeFail,
		BytesIn:    s.BytesIn,
		BytesOut:   s.BytesOut,
		Closed:     p.Closed(),
	}
}

// Run executes job with a background context.
func (p *Pipeline) Run(job Job) (*Result, error) {
	return p.RunContext(context.Background(), job)
}

// RunContext executes decode → transforms → encode, respecting cancellation.
func (p *Pipeline) RunContext(ctx context.Context, job Job) (*Result, error) {
	if p == nil {
		return nil, ErrNilPipeline
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCanceled, err)
	}
	start := time.Now()
	if err := validateJob(job, p.opts.MaxInputBytes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJob, err)
	}

	// Copy input so later mutation of caller's Raw cannot poison cache/output.
	raw := clone.Bytes(job.Raw)
	strip := job.StripExif || p.opts.StripExifByDefault
	if strip {
		raw = exifstrip.StripJPEG(raw)
	}

	key := job.CacheHint
	if key == "" {
		key = computeKey(raw, job.Transforms, job.Encode, strip)
	}

	if !job.SkipCache {
		if fr, ok := p.cache.Get(key); ok {
			p.metrics.Hit(int64(len(job.Raw)))
			out, err := p.encodeFrame(ctx, fr, job.Encode)
			if err != nil {
				p.metrics.EncodeError()
				return nil, err
			}
			return &Result{
				Bytes:    out,
				Format:   string(effectiveFormat(job.Encode, p.opts)),
				Width:    fr.W,
				Height:   fr.H,
				CacheKey: key,
				CacheHit: true,
				Duration: time.Since(start),
				Frame:    fr,
			}, nil
		}
		p.metrics.Miss(int64(len(job.Raw)))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCanceled, err)
	}

	img, format, err := decode.Decode(ctx, raw)
	if err != nil {
		p.metrics.DecodeError()
		return nil, err
	}
	maxPix := job.MaxPixels
	if maxPix <= 0 {
		maxPix = p.opts.MaxPixels
	}
	b := img.Bounds()
	if b.Dx()*b.Dy() > maxPix {
		return nil, ErrTooLarge
	}
	frame := FromImage(img, format, raw)

	for i, spec := range job.Transforms {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCanceled, err)
		}
		next, err := applyTransform(frame, spec)
		if err != nil {
			return nil, fmt.Errorf("transform[%d] %s: %w", i, spec.Kind, err)
		}
		frame = next
	}

	if !job.SkipCache {
		_ = p.cache.Put(key, frame)
	}

	out, err := p.encodeFrame(ctx, frame, job.Encode)
	if err != nil {
		p.metrics.EncodeError()
		return nil, err
	}
	p.metrics.RunOK(int64(len(job.Raw)), int64(len(out)))
	return &Result{
		Bytes:    out,
		Format:   string(effectiveFormat(job.Encode, p.opts)),
		Width:    frame.W,
		Height:   frame.H,
		CacheKey: key,
		CacheHit: false,
		Duration: time.Since(start),
		Frame:    frame,
	}, nil
}

func computeKey(raw []byte, transforms []TransformSpec, enc EncodeOptions, strip bool) string {
	specs := make([]cachekey.Spec, len(transforms))
	for i, t := range transforms {
		specs[i] = cachekey.Spec{
			Kind:    string(t.Kind),
			Width:   t.Width,
			Height:  t.Height,
			Filter:  string(t.Filter),
			X:       t.X,
			Y:       t.Y,
			Overlay: t.Overlay,
			Opacity: t.Opacity,
			Degrees: t.Degrees,
			FlipH:   t.FlipH,
			FlipV:   t.FlipV,
		}
	}
	return cachekey.Compute(raw, specs, cachekey.EncodeSpec{
		Format:  string(enc.Format),
		Quality: enc.Quality,
	}, strip)
}

func effectiveFormat(enc EncodeOptions, opts Options) EncodeFormat {
	if enc.Format != "" {
		return enc.Format
	}
	return opts.DefaultEncode
}

func (p *Pipeline) encodeFrame(ctx context.Context, fr *Frame, enc EncodeOptions) ([]byte, error) {
	if fr == nil {
		return nil, ErrInvalidFrame
	}
	format := effectiveFormat(enc, p.opts)
	quality := enc.Quality
	if quality <= 0 {
		quality = p.opts.DefaultQuality
	}
	rgba := fr.AsRGBA()
	return p.encoders.Encode(ctx, string(format), rgba, quality)
}

func applyTransform(fr *Frame, spec TransformSpec) (*Frame, error) {
	rgba := fr.AsRGBA()
	switch spec.Kind {
	case TransformScale, "":
		if spec.Width <= 0 && spec.Height <= 0 {
			return nil, ErrBadScale
		}
		filter := string(spec.Filter)
		if filter == "" {
			filter = string(FilterBilinear)
		}
		img, err := xform.Scale(rgba, spec.Width, spec.Height, filter)
		if err != nil {
			return nil, err
		}
		return FromImage(img, fr.Format, fr.Raw), nil
	case TransformCrop:
		img, err := xform.Crop(rgba, spec.X, spec.Y, spec.Width, spec.Height)
		if err != nil {
			return nil, err
		}
		return FromImage(img, fr.Format, fr.Raw), nil
	case TransformOverlay:
		img, err := xform.Overlay(rgba, spec.Overlay, spec.X, spec.Y, spec.Opacity)
		if err != nil {
			return nil, err
		}
		return FromImage(img, fr.Format, fr.Raw), nil
	case TransformRotate:
		img, err := xform.Rotate(rgba, spec.Degrees)
		if err != nil {
			return nil, err
		}
		return FromImage(img, fr.Format, fr.Raw), nil
	case TransformFlip:
		img := xform.Flip(rgba, spec.FlipH, spec.FlipV)
		return FromImage(img, fr.Format, fr.Raw), nil
	default:
		return nil, fmt.Errorf("%w: unknown kind %q", ErrInvalidJob, spec.Kind)
	}
}

// PreviewKey exposes the cache key that would be used for job without running it.
func (p *Pipeline) PreviewKey(job Job) string {
	if job.CacheHint != "" {
		return job.CacheHint
	}
	raw := clone.Bytes(job.Raw)
	strip := job.StripExif || (p != nil && p.opts.StripExifByDefault)
	if strip {
		raw = exifstrip.StripJPEG(raw)
	}
	return computeKey(raw, job.Transforms, job.Encode, strip)
}
