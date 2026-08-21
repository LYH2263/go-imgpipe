// Package imgpipe provides an image transform pipeline: decode JPEG/PNG/GIF,
// apply geometric transforms (scale, crop, overlay), encode output, and
// content-addressed disk/memory caching.
//
// Primary types are Pipeline, Frame and Job — not a generic record fan-out
// hub and not an object-store multipart session manager.
package imgpipe
