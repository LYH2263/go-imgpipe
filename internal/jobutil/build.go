package jobutil

import "github.com/LYH2263/go-imgpipe"

// ScaleJob builds a simple scale+encode job.
func ScaleJob(raw []byte, w, h int, filter imgpipe.ScaleFilter, enc imgpipe.EncodeOptions) imgpipe.Job {
        return imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{
                        Kind:   imgpipe.TransformScale,
                        Width:  w,
                        Height: h,
                        Filter: filter,
                }},
                Encode: enc,
        }
}

// ThumbnailJob scales to fit within max edge then encodes JPEG q=80.
func ThumbnailJob(raw []byte, maxEdge int) imgpipe.Job {
        return imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{
                        Kind:   imgpipe.TransformScale,
                        Width:  maxEdge,
                        Height: 0,
                        Filter: imgpipe.FilterBilinear,
                }},
                Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 80},
                StripExif: true,
        }
}

// CropEncode builds crop then encode.
func CropEncode(raw []byte, x, y, w, h int, enc imgpipe.EncodeOptions) imgpipe.Job {
        return imgpipe.Job{
                Raw: raw,
                Transforms: []imgpipe.TransformSpec{{
                        Kind:   imgpipe.TransformCrop,
                        X:      x,
                        Y:      y,
                        Width:  w,
                        Height: h,
                }},
                Encode: enc,
        }
}

// Chain appends transforms.
func Chain(job imgpipe.Job, specs ...imgpipe.TransformSpec) imgpipe.Job {
        job.Transforms = append(append([]imgpipe.TransformSpec{}, job.Transforms...), specs...)
        return job
}
