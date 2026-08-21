package validate

import "fmt"

// InputBytes checks non-empty and max size.
func InputBytes(n, max int) error {
        if n == 0 {
                return fmt.Errorf("empty input")
        }
        if max > 0 && n > max {
                return fmt.Errorf("input %d exceeds max %d", n, max)
        }
        return nil
}

// ScaleParams checks scale dimensions and filter name.
func ScaleParams(width, height int, filter string) error {
        if width < 0 || height < 0 {
                return fmt.Errorf("negative scale")
        }
        if width == 0 && height == 0 {
                return fmt.Errorf("zero scale")
        }
        switch filter {
        case "", "nearest", "bilinear":
        default:
                return fmt.Errorf("unknown filter %q", filter)
        }
        return nil
}

// CropParams checks crop rectangle.
func CropParams(x, y, w, h int) error {
        if w <= 0 || h <= 0 || x < 0 || y < 0 {
                return fmt.Errorf("invalid crop")
        }
        return nil
}

// EncodeParams checks format name / quality. Unsupported formats may still
// pass validation and fail later with ErrNoEncoder from the registry.
func EncodeParams(format string, quality int) error {
        switch format {
        case "", "jpeg", "png", "gif", "webp":
        default:
                return fmt.Errorf("unknown encode format %s", format)
        }
        if quality < 0 || quality > 100 {
                return fmt.Errorf("quality out of range")
        }
        return nil
}

// Dimensions checks W/H positive.
func Dimensions(w, h int) error {
        if w <= 0 || h <= 0 {
                return fmt.Errorf("invalid dimensions")
        }
        return nil
}

// RotateDegrees accepts 90-degree multiples.
func RotateDegrees(d int) error {
        d = d % 360
        if d < 0 {
                d += 360
        }
        if d%90 != 0 {
                return fmt.Errorf("only 90-degree multiples supported")
        }
        return nil
}
