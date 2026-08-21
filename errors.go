package imgpipe

import (
        "errors"

        "github.com/LYH2263/go-imgpipe/internal/errs"
)

var (
        ErrClosed            = errors.New("imgpipe: pipeline closed")
        ErrNilPipeline       = errors.New("imgpipe: nil pipeline")
        ErrEmptyInput        = errs.ErrEmptyInput
        ErrUnsupportedFormat = errs.ErrUnsupportedFormat
        ErrNoEncoder         = errs.ErrNoEncoder
        ErrInvalidJob        = errs.ErrInvalidJob
        ErrInvalidFrame      = errs.ErrInvalidFrame
        ErrCacheMiss         = errors.New("imgpipe: cache miss")
        ErrCacheCorrupt      = errors.New("imgpipe: cache entry corrupt")
        ErrCanceled          = errs.ErrCanceled
        ErrTooLarge          = errors.New("imgpipe: image exceeds size limit")
        ErrBadCrop           = errs.ErrBadCrop
        ErrBadScale          = errs.ErrBadScale
        ErrOverlay           = errs.ErrOverlay
)
