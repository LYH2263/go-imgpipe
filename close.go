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

	// Drop in-memory cache entries (they hold pixel buffers) but keep the
	// mem/encoders/metrics objects themselves live. A Run already past the
	// Closed gate may still dereference p.mem or p.encoders after we return;
	// nil-ing them here is the null-pointer panic through mem.Put. Clearing
	// the index frees the buffers without pulling the rug out from under any
	// in-flight request.
	p.mem.Clear()
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
