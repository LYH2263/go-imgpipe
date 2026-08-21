package imgpipe

import "time"

// Options configures a Pipeline.
type Options struct {
        // CacheDir is the on-disk content cache root. Empty disables disk persistence
        // but keeps an in-memory index.
        CacheDir string
        // MemLimit is max in-memory cache entries (default 256).
        MemLimit int
        // DiskLimitBytes caps total on-disk cache size (0 = unlimited).
        DiskLimitBytes int64
        // MaxInputBytes rejects jobs larger than this (default 32<<20).
        MaxInputBytes int
        // MaxPixels rejects decoded images with W*H above this (default 64e6).
        MaxPixels int
        // DefaultQuality for JPEG when Job.Encode.Quality is 0 (default 85).
        DefaultQuality int
        // DefaultEncode format when unset (default jpeg).
        DefaultEncode EncodeFormat
        // StripExifByDefault applies EXIF APP1 stripping before decode when true.
        StripExifByDefault bool
        // EncodeWorkers is reserved for parallel encode paths (default 1).
        EncodeWorkers int
        // SyncOnWrite forces Sync on cache temp files before rename (default true).
        SyncOnWrite bool
        // JournalName overrides the cache journal filename.
        JournalName string
        // IdleTimeout unused by library; exposed for daemon health.
        IdleTimeout time.Duration
}

func (o Options) withDefaults() Options {
        if o.MemLimit <= 0 {
                o.MemLimit = 256
        }
        if o.MaxInputBytes <= 0 {
                o.MaxInputBytes = 32 << 20
        }
        if o.MaxPixels <= 0 {
                o.MaxPixels = 64_000_000
        }
        if o.DefaultQuality <= 0 {
                o.DefaultQuality = 85
        }
        if o.DefaultEncode == "" {
                o.DefaultEncode = EncodeJPEG
        }
        if o.EncodeWorkers <= 0 {
                o.EncodeWorkers = 1
        }
        if o.JournalName == "" {
                o.JournalName = "cache.journal"
        }
        if o.IdleTimeout <= 0 {
                o.IdleTimeout = 30 * time.Second
        }
        // SyncOnWrite defaults true; callers must set a sentinel to disable.
        // We treat zero-value as true via explicit flag below in New.
        return o
}
