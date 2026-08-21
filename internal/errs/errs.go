package errs

import "errors"

var (
        ErrUnsupportedFormat = errors.New("imgpipe: unsupported image format")
        ErrNoEncoder         = errors.New("imgpipe: no encoder registered for format")
        ErrCanceled          = errors.New("imgpipe: canceled")
        ErrEmptyInput        = errors.New("imgpipe: empty input")
        ErrInvalidFrame      = errors.New("imgpipe: invalid frame")
        ErrBadCrop           = errors.New("imgpipe: invalid crop rectangle")
        ErrBadScale          = errors.New("imgpipe: invalid scale parameters")
        ErrOverlay           = errors.New("imgpipe: overlay failed")
        ErrInvalidJob        = errors.New("imgpipe: invalid job")
)
