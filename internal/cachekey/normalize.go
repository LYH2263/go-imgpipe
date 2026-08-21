package cachekey

import "strings"

// NormalizeFormat lowercases and maps aliases to canonical names.
func NormalizeFormat(s string) string {
        s = strings.ToLower(strings.TrimSpace(s))
        switch s {
        case "jpg", "jpeg", "image/jpeg":
                return "jpeg"
        case "png", "image/png":
                return "png"
        case "gif", "image/gif":
                return "gif"
        default:
                return s
        }
}

// CanonicalTransformKind normalizes kind aliases.
func CanonicalTransformKind(s string) string {
        s = strings.ToLower(strings.TrimSpace(s))
        switch s {
        case "resize", "scale", "":
                return "scale"
        case "crop", "clip":
                return "crop"
        case "watermark", "overlay":
                return "overlay"
        case "rotate", "rot":
                return "rotate"
        case "flip", "mirror":
                return "flip"
        default:
                return s
        }
}
