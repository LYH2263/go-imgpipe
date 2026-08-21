package jobutil

import (
        "encoding/json"
        "fmt"
        "strconv"
        "strings"

        "github.com/LYH2263/go-imgpipe"
        "github.com/LYH2263/go-imgpipe/internal/cachekey"
)

// TransformQuery parses strings like "scale:200x100:bilinear" or "crop:10,20,100,80".
func TransformQuery(s string) (imgpipe.TransformSpec, error) {
        s = strings.TrimSpace(s)
        if s == "" {
                return imgpipe.TransformSpec{}, fmt.Errorf("empty transform")
        }
        parts := strings.SplitN(s, ":", 3)
        kind := cachekey.CanonicalTransformKind(parts[0])
        spec := imgpipe.TransformSpec{Kind: imgpipe.TransformKind(kind)}
        switch kind {
        case "scale":
                if len(parts) < 2 {
                        return spec, fmt.Errorf("scale needs WxH")
                }
                wh := strings.Split(parts[1], "x")
                if len(wh) != 2 {
                        return spec, fmt.Errorf("bad scale size")
                }
                w, err1 := strconv.Atoi(wh[0])
                h, err2 := strconv.Atoi(wh[1])
                if err1 != nil || err2 != nil {
                        return spec, fmt.Errorf("bad scale ints")
                }
                spec.Width, spec.Height = w, h
                if len(parts) == 3 {
                        spec.Filter = imgpipe.ScaleFilter(parts[2])
                }
        case "crop":
                if len(parts) < 2 {
                        return spec, fmt.Errorf("crop needs x,y,w,h")
                }
                nums := strings.Split(parts[1], ",")
                if len(nums) != 4 {
                        return spec, fmt.Errorf("crop needs 4 ints")
                }
                vals := make([]int, 4)
                for i, n := range nums {
                        v, err := strconv.Atoi(strings.TrimSpace(n))
                        if err != nil {
                                return spec, err
                        }
                        vals[i] = v
                }
                spec.X, spec.Y, spec.Width, spec.Height = vals[0], vals[1], vals[2], vals[3]
        case "rotate":
                if len(parts) < 2 {
                        return spec, fmt.Errorf("rotate needs degrees")
                }
                d, err := strconv.Atoi(parts[1])
                if err != nil {
                        return spec, err
                }
                spec.Degrees = d
        case "flip":
                if len(parts) >= 2 {
                        switch strings.ToLower(parts[1]) {
                        case "h", "horizontal":
                                spec.FlipH = true
                        case "v", "vertical":
                                spec.FlipV = true
                        case "both":
                                spec.FlipH, spec.FlipV = true, true
                        }
                } else {
                        spec.FlipH = true
                }
        default:
                return spec, fmt.Errorf("unknown transform %q", kind)
        }
        return spec, nil
}

// ParseTransformList splits on '|' and parses each.
func ParseTransformList(s string) ([]imgpipe.TransformSpec, error) {
        if strings.TrimSpace(s) == "" {
                return nil, nil
        }
        parts := strings.Split(s, "|")
        out := make([]imgpipe.TransformSpec, 0, len(parts))
        for _, p := range parts {
                t, err := TransformQuery(p)
                if err != nil {
                        return nil, err
                }
                out = append(out, t)
        }
        return out, nil
}

// JobJSON is a wire format for the HTTP API (raw provided separately).
type JobJSON struct {
        Transforms []imgpipe.TransformSpec `json:"transforms"`
        Encode     imgpipe.EncodeOptions   `json:"encode"`
        SkipCache  bool                    `json:"skip_cache"`
        StripExif  bool                    `json:"strip_exif"`
        CacheHint  string                  `json:"cache_hint"`
}

func ParseJobJSON(b []byte) (JobJSON, error) {
        var j JobJSON
        if err := json.Unmarshal(b, &j); err != nil {
                return j, err
        }
        return j, nil
}
