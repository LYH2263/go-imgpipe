package validate

// ClampQuality returns quality in [1,100], defaulting zero to def.
func ClampQuality(q, def int) int {
        if q <= 0 {
                q = def
        }
        if q < 1 {
                q = 1
        }
        if q > 100 {
                q = 100
        }
        return q
}

// FitWithin scales w,h to fit inside maxW,maxH preserving aspect. Zero max means unlimited.
func FitWithin(w, h, maxW, maxH int) (int, int) {
        if w <= 0 || h <= 0 {
                return w, h
        }
        if maxW <= 0 && maxH <= 0 {
                return w, h
        }
        nw, nh := w, h
        if maxW > 0 && nw > maxW {
                nh = nh * maxW / nw
                nw = maxW
        }
        if maxH > 0 && nh > maxH {
                nw = nw * maxH / nh
                nh = maxH
        }
        if nw < 1 {
                nw = 1
        }
        if nh < 1 {
                nh = 1
        }
        return nw, nh
}

// PixelCount returns w*h with overflow guard.
func PixelCount(w, h int) int {
        if w <= 0 || h <= 0 {
                return 0
        }
        if w > 0 && h > (1<<31-1)/w {
                return 1<<31 - 1
        }
        return w * h
}
