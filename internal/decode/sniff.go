package decode

import "github.com/LYH2263/go-imgpipe/internal/errs"

// Sniff returns FormatJPEG/PNG/GIF or empty string.
func Sniff(data []byte) string {
        if len(data) < 3 {
                return ""
        }
        if data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
                return errs.FormatJPEG
        }
        if len(data) >= 8 &&
                data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 &&
                data[4] == 0x0d && data[5] == 0x0a && data[6] == 0x1a && data[7] == 0x0a {
                return errs.FormatPNG
        }
        if len(data) >= 6 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F' &&
                data[3] == '8' && (data[4] == '7' || data[4] == '9') && data[5] == 'a' {
                return errs.FormatGIF
        }
        return ""
}

func IsJPEG(data []byte) bool { return Sniff(data) == errs.FormatJPEG }
func IsPNG(data []byte) bool  { return Sniff(data) == errs.FormatPNG }
func IsGIF(data []byte) bool  { return Sniff(data) == errs.FormatGIF }
