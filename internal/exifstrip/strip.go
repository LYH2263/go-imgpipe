package exifstrip

import "github.com/LYH2263/go-imgpipe/internal/clone"

// StripJPEG removes APP1 (EXIF) and APP2 segments from a JPEG bitstream.
// Non-JPEG input is returned unchanged (copied).
func StripJPEG(data []byte) []byte {
        if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
                return clone.Bytes(data)
        }
        out := make([]byte, 0, len(data))
        out = append(out, 0xff, 0xd8)
        i := 2
        for i < len(data) {
                if data[i] != 0xff {
                        // raw entropy-coded data: copy rest
                        out = append(out, data[i:]...)
                        break
                }
                // skip fill bytes
                for i < len(data) && data[i] == 0xff {
                        i++
                }
                if i >= len(data) {
                        break
                }
                marker := data[i]
                i++
                if marker == 0xd9 { // EOI
                        out = append(out, 0xff, 0xd9)
                        break
                }
                if marker == 0xda { // SOS — copy rest including marker
                        out = append(out, 0xff, marker)
                        out = append(out, data[i:]...)
                        break
                }
                // markers without length
                if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
                        out = append(out, 0xff, marker)
                        continue
                }
                if i+1 >= len(data) {
                        break
                }
                seglen := int(data[i])<<8 | int(data[i+1])
                if seglen < 2 || i+seglen > len(data) {
                        out = append(out, data[i-2:]...)
                        break
                }
                // drop APP1 (E1) and APP2 (E2)
                if marker == 0xe1 || marker == 0xe2 {
                        i += seglen
                        continue
                }
                out = append(out, 0xff, marker)
                out = append(out, data[i:i+seglen]...)
                i += seglen
        }
        return out
}

// HasAPP1 reports whether JPEG contains an APP1 segment.
func HasAPP1(data []byte) bool {
        if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
                return false
        }
        i := 2
        for i+3 < len(data) {
                if data[i] != 0xff {
                        return false
                }
                for i < len(data) && data[i] == 0xff {
                        i++
                }
                if i >= len(data) {
                        return false
                }
                marker := data[i]
                i++
                if marker == 0xda || marker == 0xd9 {
                        return false
                }
                if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
                        continue
                }
                if i+1 >= len(data) {
                        return false
                }
                seglen := int(data[i])<<8 | int(data[i+1])
                if seglen < 2 || i+seglen > len(data) {
                        return false
                }
                if marker == 0xe1 {
                        return true
                }
                i += seglen
        }
        return false
}
