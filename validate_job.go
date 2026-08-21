package imgpipe

import (
	"fmt"

	"github.com/LYH2263/go-imgpipe/internal/validate"
)

func validateJob(job Job, maxInput int) error {
	if err := validate.InputBytes(len(job.Raw), maxInput); err != nil {
		return err
	}
	if job.MaxPixels < 0 {
		return fmt.Errorf("max pixels negative")
	}
	for i, t := range job.Transforms {
		if err := validateTransform(t); err != nil {
			return fmt.Errorf("transform[%d]: %w", i, err)
		}
	}
	return validate.EncodeParams(string(job.Encode.Format), job.Encode.Quality)
}

func validateTransform(t TransformSpec) error {
	switch t.Kind {
	case TransformScale, "":
		return validate.ScaleParams(t.Width, t.Height, string(t.Filter))
	case TransformCrop:
		return validate.CropParams(t.X, t.Y, t.Width, t.Height)
	case TransformOverlay:

		return nil
	case TransformRotate:
		return validate.RotateDegrees(t.Degrees)
	case TransformFlip:
		return nil
	default:
		return fmt.Errorf("unknown kind %q", t.Kind)
	}
}
