package imgpipe

import (
        "image/color"
        "time"
)

// Format names used by the pipeline and encoders.
const (
        FormatJPEG = "jpeg"
        FormatPNG  = "png"
        FormatGIF  = "gif"
        FormatRaw  = "raw"
)

// TransformKind selects a geometric or compositing step.
type TransformKind string

const (
        TransformScale   TransformKind = "scale"
        TransformCrop    TransformKind = "crop"
        TransformOverlay TransformKind = "overlay"
        TransformRotate  TransformKind = "rotate"
        TransformFlip    TransformKind = "flip"
)

// ScaleFilter selects resampling algorithm.
type ScaleFilter string

const (
        FilterNearest  ScaleFilter = "nearest"
        FilterBilinear ScaleFilter = "bilinear"
)

// EncodeFormat selects output codec.
type EncodeFormat string

const (
        EncodeJPEG EncodeFormat = "jpeg"
        EncodePNG  EncodeFormat = "png"
)

// TransformSpec describes one transform in a Job chain.
type TransformSpec struct {
        Kind   TransformKind
        Width  int
        Height int
        Filter ScaleFilter
        X      int
        Y      int
        // Overlay carries opaque PNG/JPEG bytes for watermark compositing.
        Overlay []byte
        Opacity float64
        Degrees int
        FlipH   bool
        FlipV   bool
}

// EncodeOptions controls output encoding.
type EncodeOptions struct {
        Format  EncodeFormat
        Quality int // JPEG quality 1-100; ignored for PNG
}

// Job is one pipeline run: raw image bytes plus transform chain and encode opts.
type Job struct {
        Raw         []byte
        Transforms  []TransformSpec
        Encode      EncodeOptions
        SkipCache   bool
        StripExif   bool
        CacheHint   string
        MaxPixels   int
}

// Result is the encoded output plus cache metadata.
type Result struct {
        Bytes      []byte
        Format     string
        Width      int
        Height     int
        CacheKey   string
        CacheHit   bool
        Duration   time.Duration
        Frame      *Frame
}

// Stats snapshots pipeline counters.
type Stats struct {
        Runs       int64
        CacheHits  int64
        CacheMiss  int64
        DecodeFail int64
        EncodeFail int64
        BytesIn    int64
        BytesOut   int64
        Closed     bool
}

// ColorModelName maps Frame.ColorModel to a stable string for cache keys.
func ColorModelName(m color.Model) string {
        switch m {
        case color.RGBAModel:
                return "rgba"
        case color.NRGBAModel:
                return "nrgba"
        case color.GrayModel:
                return "gray"
        case color.Gray16Model:
                return "gray16"
        case color.YCbCrModel:
                return "ycbcr"
        default:
                return "other"
        }
}
