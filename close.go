package imgpipe

import (
        "sync/atomic"
)

// Close releases disk cache resources and marks the pipeline unusable.
// It syncs the journal before releasing the mount (defer-order sensitive).
func (p *Pipeline) Close() error {
        if p == nil {
                return ErrNilPipeline
        }
        if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
                return ErrClosed
        }
        p.mu.Lock()
        defer p.mu.Unlock()
        var first error
        if p.disk != nil {
                if err := p.disk.Sync(); err != nil && first == nil {
                        first = err
                }
                if err := p.disk.Close(); err != nil && first == nil {
                        first = err
                }
                p.disk = nil
        }
        if p.mem != nil {
                p.mem.Clear()
        }
        p.metrics.SetClosed(true)
        return first
}

// Closed reports whether Close has been called.
func (p *Pipeline) Closed() bool {
        if p == nil {
                return true
        }
        return atomic.LoadInt32(&p.closed) != 0
}
