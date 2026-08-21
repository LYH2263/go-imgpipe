package encode

import (
        "strings"

        "github.com/LYH2263/go-imgpipe/internal/errs"
)

func ContentType(format string) string {
        switch strings.ToLower(format) {
        case errs.FormatJPEG, "jpg":
                return "image/jpeg"
        case errs.FormatPNG:
                return "image/png"
        case errs.FormatGIF:
                return "image/gif"
        default:
                return "application/octet-stream"
        }
}

func Ext(format string) string {
        switch strings.ToLower(format) {
        case errs.FormatJPEG, "jpg":
                return ".jpg"
        case errs.FormatPNG:
                return ".png"
        case errs.FormatGIF:
                return ".gif"
        default:
                return ".bin"
        }
}

func Supported() []string {
        return []string{errs.FormatJPEG, errs.FormatPNG}
}
