package exifstrip

// Policy controls which metadata segments are removed.
type Policy struct {
        StripAPP1 bool
        StripAPP2 bool
        StripCOM  bool
}

// DefaultPolicy strips EXIF and ICC APP segments.
func DefaultPolicy() Policy {
        return Policy{StripAPP1: true, StripAPP2: true, StripCOM: false}
}

// Apply strips segments according to policy. Non-JPEG copied unchanged.
func Apply(data []byte, p Policy) []byte {
        if !p.StripAPP1 && !p.StripAPP2 && !p.StripCOM {
                out := make([]byte, len(data))
                copy(out, data)
                return out
        }
        // reuse StripJPEG for APP1/APP2; COM handled below if needed
        out := StripJPEG(data)
        if !p.StripCOM {
                return out
        }
        return stripCOM(out)
}

func stripCOM(data []byte) []byte {
        if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
                return data
        }
        out := make([]byte, 0, len(data))
        out = append(out, 0xff, 0xd8)
        i := 2
        for i < len(data) {
                if data[i] != 0xff {
                        out = append(out, data[i:]...)
                        break
                }
                for i < len(data) && data[i] == 0xff {
                        i++
                }
                if i >= len(data) {
                        break
                }
                marker := data[i]
                i++
                if marker == 0xd9 {
                        out = append(out, 0xff, 0xd9)
                        break
                }
                if marker == 0xda {
                        out = append(out, 0xff, marker)
                        out = append(out, data[i:]...)
                        break
                }
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
                if marker == 0xfe { // COM
                        i += seglen
                        continue
                }
                out = append(out, 0xff, marker)
                out = append(out, data[i:i+seglen]...)
                i += seglen
        }
        return out
}
