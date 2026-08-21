package decode_test

import (
        "bytes"
        "context"
        "image"
        "image/color"
        "image/jpeg"
        "testing"

        "github.com/LYH2263/go-imgpipe/internal/decode"
)

func TestSniffAndDecode(t *testing.T) {
        img := image.NewRGBA(image.Rect(0, 0, 8, 8))
        img.Set(0, 0, color.RGBA{R: 1, A: 255})
        var buf bytes.Buffer
        _ = jpeg.Encode(&buf, img, nil)
        if !decode.IsJPEG(buf.Bytes()) {
                t.Fatal("sniff")
        }
        out, format, err := decode.Decode(context.Background(), buf.Bytes())
        if err != nil || format != "jpeg" {
                t.Fatalf("%s %v", format, err)
        }
        if out.Bounds().Dx() != 8 {
                t.Fatal(out.Bounds())
        }
}
